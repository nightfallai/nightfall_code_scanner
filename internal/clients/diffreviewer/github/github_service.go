package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/google/go-github/v31/github"
	"github.com/watchtowerai/nightfall_dlp/internal/clients/diffreviewer"
	"github.com/watchtowerai/nightfall_dlp/internal/interfaces"
	"github.com/watchtowerai/nightfall_dlp/internal/nightfallconfig"
)

type Level string

const (
	InfoLevel    Level = "info"
	WarningLevel Level = "warning"
	ErrorLevel   Level = "error"

	WorkspacePathEnvVar   = "GITHUB_WORKSPACE"
	EventPathEnvVar       = "GITHUB_EVENT_PATH"
	NightfallAPIKeyEnvVar = "NIGHTFALL_API_KEY"
)

var rawOptionsTypeDiff = github.RawOptions{Type: github.Diff}

type ownerLogin struct {
	Login string `json:"login"`
}

type eventRepo struct {
	Owner ownerLogin `json:"owner"`
	Name  string     `json:"name"`
}

type checkSuite struct {
	After        string        `json:"after"`
	PullRequests []pullRequest `json:"pull_requests"`
}

type headCommit struct {
	ID string `json:"id"`
}

// event represents github event webhook file
type event struct {
	PullRequest pullRequest `json:"pull_request"`
	Repository  eventRepo   `json:"repository"`
	CheckSuite  checkSuite  `json:"check_suite"`
	HeadCommit  headCommit  `json:"head_commit"`
}

type ownerID struct {
	ID int64 `json:"id"`
}

// repo contains information relevant to the
// github repo being checked
type repo struct {
	Owner ownerID `json:"owner"`
}

type pullRequestHead struct {
	Sha  string `json:"sha"`
	Ref  string `json:"ref"`
	Repo repo   `json:"repo"`
}

type pullRequestBase struct {
	Repo repo `json:"repo"`
}

// pullRequest contains information relevant to
// the github pull request
type pullRequest struct {
	Number int             `json:"number"`
	Head   pullRequestHead `json:"head"`
	Base   pullRequestBase `json:"base"`
}

// CheckRequest represents nightfallDLP GitHub check request.
type CheckRequest struct {
	// Commit SHA.
	// Required.
	SHA string `json:"sha,omitempty"`
	// PullRequest number.
	// Optional.
	PullRequest int `json:"pull_request,omitempty"`
	// Owner of the repository.
	// Required.
	Owner string `json:"owner,omitempty"`
	// Repository name.
	// Required.
	Repo string `json:"repo,omitempty"`

	// Annotations associated with the repository's commit and Pull Request.
	Annotations []*Annotation `json:"annotations,omitempty"`

	// Name of the annotation tool.
	// Optional.
	Name string `json:"name,omitempty"`

	// Level is report level for this request.
	// One of ["info", "warning", "error"]. Default is "error".
	// Optional.
	Level Level `json:"level"`

	// FilterMode represents a way to filter checks results
	// Optional. TODO check to see if this is necessary
	// FilterMode difffilter.Mode `json:"filter_mode"`
}

// CheckResponse represents nightfallDLP GitHub check response.
type CheckResponse struct {
	// ReportURL is report URL of check run.
	ReportURL string `json:"report_url,omitempty"`
	// CheckedResults is checked annotations result.
	CheckedResults []*Annotation `json:"checked_results"`
	// Conclusion of check result https://developer.github.com/v3/checks/runs/#parameters-1
	Conclusion string `json:"conclusion,omitempty"`
}

// Annotation represents an annotation to file or specific line.
// https://developer.github.com/v3/checks/runs/#annotations-object
type Annotation struct {
	Path       string `json:"path,omitempty"`
	Line       int    `json:"line,omitempty"`
	Message    string `json:"message,omitempty"`
	RawMessage string `json:"raw_message,omitempty"`
}

// Service contains the github client that makes Github api calls
type Service struct {
	Client       interfaces.GithubAPI
	CheckRequest *CheckRequest
	CheckRun     *github.CheckRun
}

// NewGithubService create a new github service
func NewGithubService(httpClientInterface interfaces.HTTPClient) diffreviewer.DiffReviewer {
	githubClient := NewClient(httpClientInterface)
	return &Service{
		Client: githubClient,
	}
}

func getEventFile(eventPath string) (*event, error) {
	f, err := os.Open(eventPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var event event
	if err := json.NewDecoder(f).Decode(&event); err != nil {
		return nil, err
	}
	return &event, nil
}

// LoadConfig gets all config values from files or environment and creates a config
func (s *Service) LoadConfig(nightfallConfigFileName string) (*nightfallconfig.Config, error) {
	workspacePath, ok := os.LookupEnv(WorkspacePathEnvVar)
	if !ok {
		return nil, errors.New("Missing env var for workspace path")
	}
	eventPath, ok := os.LookupEnv(EventPathEnvVar)
	if !ok {
		return nil, errors.New("Missing env var for event path")
	}
	event, err := getEventFile(eventPath)
	if err != nil {
		return nil, err
	}
	s.CheckRequest = &CheckRequest{
		Owner:       event.Repository.Owner.Login,
		Repo:        event.Repository.Name,
		SHA:         event.PullRequest.Head.Sha,
		PullRequest: event.PullRequest.Number,
	}
	if s.CheckRequest.SHA == "" {
		s.CheckRequest.SHA = event.HeadCommit.ID
	}
	nightfallConfig, err := nightfallconfig.GetConfigFile(workspacePath, nightfallConfigFileName)
	if err != nil {
		return nil, err
	}
	nightfallAPIKey, ok := os.LookupEnv(NightfallAPIKeyEnvVar)
	if !ok {
		return nil, errors.New("Missing env var for nightfall api key")
	}
	return &nightfallconfig.Config{
		NightfallAPIKey:    nightfallAPIKey,
		NightfallDetectors: nightfallConfig.Detectors,
	}, nil
}

// GetDiff retrieves the file diff from the requested pull request
func (s *Service) GetDiff() ([]*diffreviewer.FileDiff, error) {
	// TODO look into how we can retrieve the diff through the github action yaml file
	ctx := context.Background()
	d, _, err := s.Client.GetRaw(
		ctx,
		s.CheckRequest.Owner,
		s.CheckRequest.Repo,
		s.CheckRequest.PullRequest,
		rawOptionsTypeDiff,
	)
	if err != nil {
		return nil, err
	}
	fileDiffs, err := ParseMultiFile(bytes.NewReader([]byte(d)))
	if err != nil {
		return nil, err
	}
	fileDiffs = filterFileDiffs(fileDiffs)
	return fileDiffs, nil
}

func filterFileDiffs(fileDiffs []*diffreviewer.FileDiff) []*diffreviewer.FileDiff {
	if len(fileDiffs) == 0 {
		return fileDiffs
	}
	filteredFileDiffs := []*diffreviewer.FileDiff{}
	for _, fileDiff := range fileDiffs {
		fileDiff.Hunks = filterHunks(fileDiff.Hunks)
		if len(fileDiff.Hunks) > 0 {
			filteredFileDiffs = append(filteredFileDiffs, fileDiff)
		}
	}
	return filteredFileDiffs
}

func filterHunks(hunks []*diffreviewer.Hunk) []*diffreviewer.Hunk {
	filteredHunks := []*diffreviewer.Hunk{}
	for _, hunk := range hunks {
		hunk.Lines = filterLines(hunk.Lines)
		if len(hunk.Lines) > 0 {
			filteredHunks = append(filteredHunks, hunk)
		}
	}
	return filteredHunks
}

func filterLines(lines []*diffreviewer.Line) []*diffreviewer.Line {
	filteredLines := []*diffreviewer.Line{}
	for _, line := range lines {
		if line.Type == diffreviewer.LineAdded {
			filteredLines = append(filteredLines, line)
		}
	}
	return filteredLines
}

// WriteComments posts the findings as annotations to the github check
func (s *Service) WriteComments(
	comments []*diffreviewer.Comment,
) error {
	//TODO implement
	return nil
}

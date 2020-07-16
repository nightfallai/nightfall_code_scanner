package github

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"

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
	RepoEnvVar            = "GITHUB_REPOSITORY"
	CommitShaEnvVar       = "GITHUB_SHA"
	NightfallAPIKeyEnvVar = "NIGHTFALL_API_KEY"
)

var rawOptionsTypeDiff = github.RawOptions{Type: github.Diff}

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

// Client contains the github client that makes Github api calls
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

// LoadConfig gets all config values from files or environment and creates a config
func (s *Service) LoadConfig(nightfallConfigFileName string) (*nightfallconfig.Config, error) {
	workspacePath, ok := os.LookupEnv(WorkspacePathEnvVar)
	if !ok {
		return nil, errors.New("Missing env var for workspace path")
	}
	nightfallConfig, err := nightfallconfig.GetConfigFile(workspacePath, nightfallConfigFileName)
	if err != nil {
		return nil, err
	}
	repoFullName, ok := os.LookupEnv(RepoEnvVar)
	if !ok {
		return nil, errors.New("Missing env var for repo name")
	}
	// Format of repoFullName is <owner>/<repo_name>
	repoFullNameSplit := strings.Split(repoFullName, "/")
	if len(repoFullNameSplit) != 2 {
		return nil, errors.New("Received invalid format for repo full name")
	}
	owner := repoFullNameSplit[0]
	repo := repoFullNameSplit[1]
	commitSha, ok := os.LookupEnv(CommitShaEnvVar)
	if !ok {
		return nil, errors.New("Missing env var for repo name")
	}
	nightfallAPIKey, ok := os.LookupEnv(NightfallAPIKeyEnvVar)
	if !ok {
		return nil, errors.New("Missing env var for nightfall api key")
	}
	s.CheckRequest = &CheckRequest{
		Owner: owner,
		Repo:  repo,
		SHA:   commitSha,
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

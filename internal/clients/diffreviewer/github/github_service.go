package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
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

	WorkspacePathEnvVar    = "GITHUB_WORKSPACE"
	EventPathEnvVar        = "GITHUB_EVENT_PATH"
	NightfallAPIKeyEnvVar  = "NIGHTFALL_API_KEY"
	GithubBaseBranchEnvVar = "NIGHTFALL_GITHUB_BASE_BRANCH"

	MaxAnnotationsPerRequest = 50 // https://developer.github.com/v3/checks/runs/#output-object

	nightfallGithubWorkflowName = "nightfalldlp"
	githubActionAppName         = "github actions"
)

var rawOptionsTypeDiff = github.RawOptions{Type: github.Diff}
var annotationLevelFailure = "failure"
var checkRunCompletedStatus = "completed"
var checkRunInProgressStatus = "in_progress"
var checkRunConclusionSuccess = "success"
var checkRunConclusionFailure = "failure"
var defaultBaseBranch = "master"
var checkRunOptsInProgressStatus = github.ListCheckRunsOptions{Status: &checkRunInProgressStatus}

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
	BaseBranch   string
}

// NewGithubService creates a new github service with the given httpClient
func NewGithubService(httpClientInterface interfaces.HTTPClient) diffreviewer.DiffReviewer {
	githubClient := NewClient(httpClientInterface)
	return &Service{
		Client: githubClient,
	}
}

// NewAuthenticatedGithubService creates a new authenticated github service with the github token
func NewAuthenticatedGithubService(githubToken string) diffreviewer.DiffReviewer {
	githubClient := NewAuthenticatedClient(githubToken)
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
	if baseBranch, ok := os.LookupEnv(GithubBaseBranchEnvVar); ok {
		s.BaseBranch = baseBranch
	} else {
		s.BaseBranch = defaultBaseBranch
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
	var d string
	var err error
	if s.CheckRequest.PullRequest == 0 {
		d, _, err = s.Client.GetRawBySha(
			ctx,
			s.CheckRequest.Owner,
			s.CheckRequest.Repo,
			s.BaseBranch,
			s.CheckRequest.SHA,
		)
	} else {
		d, _, err = s.Client.GetRaw(
			ctx,
			s.CheckRequest.Owner,
			s.CheckRequest.Repo,
			s.CheckRequest.PullRequest,
			rawOptionsTypeDiff,
		)
	}
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
	checkRun, err := s.getCheckRun()
	if err != nil {
		return err
	}
	conclusionStatus := checkRunConclusionFailure
	if len(comments) == 0 {
		conclusionStatus = checkRunConclusionSuccess
	}
	numUpdateRequests := int(math.Ceil(float64(len(comments)) / MaxAnnotationsPerRequest))
	for i := 0; i < numUpdateRequests; i++ {
		startCommentIdx := i * MaxAnnotationsPerRequest
		endCommentIdx := min(startCommentIdx+MaxAnnotationsPerRequest, len(comments))
		annotations := createAnnotations(comments[startCommentIdx:endCommentIdx])
		opt := github.UpdateCheckRunOptions{
			Name: getCheckName(s.CheckRequest.Name),
			Output: &github.CheckRunOutput{
				Title:       github.String(getCheckName(s.CheckRequest.Name)),
				Annotations: annotations,
				Summary:     github.String(""),
			},
		}
		_, _, err := s.Client.UpdateCheckRun(context.Background(),
			s.CheckRequest.Owner,
			s.CheckRequest.Repo,
			checkRun.GetID(),
			opt,
		)
		if err != nil {
			log.Printf("error posting batch #%d of comments: %s", i, err)
		}
	}
	completedOpt := github.UpdateCheckRunOptions{
		Status:     &checkRunCompletedStatus,
		Conclusion: &conclusionStatus,
	}
	_, _, err = s.Client.UpdateCheckRun(context.Background(),
		s.CheckRequest.Owner,
		s.CheckRequest.Repo,
		checkRun.GetID(),
		completedOpt,
	)
	if err != nil {
		return err
	}
	return nil
}

// getCheckRun gets the github check run associated with the github action
func (s *Service) getCheckRun() (*github.CheckRun, error) {
	listCheckRuns, _, err := s.Client.ListCheckRunsForRef(
		context.Background(),
		s.CheckRequest.Owner,
		s.CheckRequest.Repo,
		s.CheckRequest.SHA,
		&checkRunOptsInProgressStatus,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get check run: %v", err)
	}
	for _, run := range listCheckRuns.CheckRuns {
		runNameLow := strings.ToLower(run.GetName())
		appNameLow := strings.ToLower(run.App.GetName())
		if strings.Contains(runNameLow, nightfallGithubWorkflowName) && appNameLow == githubActionAppName {
			return run, nil
		}
	}
	log.Println("Please check that your NightfallDLP job has the 'name' field filled as 'nightfalldlp'")
	return nil, fmt.Errorf("failed to get check run from list of check runs")
}

func createAnnotations(comments []*diffreviewer.Comment) []*github.CheckRunAnnotation {
	annotations := make([]*github.CheckRunAnnotation, len(comments))
	for i := 0; i < len(comments); i++ {
		annotations[i] = convertCommentToAnnotation(comments[i])
	}
	return annotations
}

func convertCommentToAnnotation(comment *diffreviewer.Comment) *github.CheckRunAnnotation {
	return &github.CheckRunAnnotation{
		Path:            &comment.FilePath,
		StartLine:       &comment.LineNumber,
		EndLine:         &comment.LineNumber,
		Message:         &comment.Body,
		AnnotationLevel: &annotationLevelFailure,
	}
}

func getCheckName(name string) string {
	if name != "" {
		return name
	}
	return "NightfallDLP"
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

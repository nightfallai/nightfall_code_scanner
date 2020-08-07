package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"

	"github.com/google/go-github/v31/github"
	"github.com/nightfallai/jenkins_test/internal/clients/diffreviewer"
	"github.com/nightfallai/jenkins_test/internal/clients/logger"
	githublogger "github.com/nightfallai/jenkins_test/internal/clients/logger/github_logger"
	"github.com/nightfallai/jenkins_test/internal/interfaces/githubintf"
	"github.com/nightfallai/jenkins_test/internal/nightfallconfig"
)

type Level string

const (
	InfoLevel    Level = "info"
	WarningLevel Level = "warning"
	ErrorLevel   Level = "error"

	WorkspacePathEnvVar      = "GITHUB_WORKSPACE"
	EventPathEnvVar          = "GITHUB_EVENT_PATH"
	NightfallAPIKeyEnvVar    = "NIGHTFALL_API_KEY"
	NightfallDiffFileName    = "./nightfalldlp_raw_diff.txt"
	MaxAnnotationsPerRequest = 50 // https://developer.github.com/v3/checks/runs/#output-object

	imageURL      = "https://www.finsmes.com/wp-content/uploads/2019/11/Nightfall-AI.png"
	imageAlt      = "Nightfall Logo"
	summaryString = "Nightfall DLP has found %d potentially sensitive items"
)

var annotationLevelFailure = "failure"
var checkRunCompletedStatus = "completed"
var checkRunInProgressStatus = "in_progress"
var checkRunConclusionSuccess = "success"
var checkRunConclusionFailure = "failure"

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

// CheckRequest represents a nightfallDLP GitHub check request.
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

	// Name of the annotation tool.
	// Optional.
	Name string `json:"name,omitempty"`
}

// Service contains the github client that makes Github api calls
type Service struct {
	Client       githubintf.GithubClient
	Logger       logger.Logger
	CheckRequest *CheckRequest
}

// NewAuthenticatedGithubService creates a new authenticated github service with the github token
func NewAuthenticatedGithubService(githubToken string) diffreviewer.DiffReviewer {
	return &Service{
		Client: NewAuthenticatedClient(githubToken),
		Logger: githublogger.NewDefaultGithubLogger(),
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

// GetLogger gets the github service logger
func (s *Service) GetLogger() logger.Logger {
	return s.Logger
}

// LoadConfig gets all config values from files or environment and creates a config
func (s *Service) LoadConfig(nightfallConfigFileName string) (*nightfallconfig.Config, error) {
	s.Logger.Debug("Loading configuration")
	workspacePath, ok := os.LookupEnv(WorkspacePathEnvVar)
	if !ok {
		s.Logger.Error(fmt.Sprintf("Environment variable %s cannot be found", WorkspacePathEnvVar))
		return nil, errors.New("Missing env var for workspace path")
	}
	eventPath, ok := os.LookupEnv(EventPathEnvVar)
	if !ok {
		s.Logger.Error(fmt.Sprintf("Environment variable %s cannot be found", EventPathEnvVar))
		return nil, errors.New("Missing env var for event path")
	}
	event, err := getEventFile(eventPath)
	if err != nil {
		s.Logger.Error("Error getting Github event file")
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
		s.Logger.Error("Error getting Nightfall config file. Ensure you have a Nightfall config file located in the root of your repository at .nightfalldlp/config.json with at least one Detector enabled")
		return nil, err
	}
	nightfallAPIKey, ok := os.LookupEnv(NightfallAPIKeyEnvVar)
	if !ok {
		s.Logger.Error(fmt.Sprintf("Error getting Nightfall API key. Ensure you have %s set in the Github secrets of the repo", NightfallAPIKeyEnvVar))
		return nil, errors.New("Missing env var for nightfall api key")
	}
	return &nightfallconfig.Config{
		NightfallAPIKey:            nightfallAPIKey,
		NightfallDetectors:         nightfallConfig.Detectors,
		NightfallMaxNumberRoutines: nightfallConfig.MaxNumberRoutines,
		TokenExclusionList:         nightfallConfig.TokenExclusionList,
		FileInclusionList:          nightfallConfig.FileInclusionList,
		FileExclusionList:          nightfallConfig.FileExclusionList,
	}, nil
}

// GetDiff retrieves the file diff from the requested pull request
func (s *Service) GetDiff() ([]*diffreviewer.FileDiff, error) {
	s.Logger.Debug("Getting diff from Github")
	content, err := ioutil.ReadFile(NightfallDiffFileName)
	if err != nil {
		s.Logger.Error("Error getting the raw diff from Github")
		return nil, err
	}
	fileDiffs, err := ParseMultiFile(bytes.NewReader(content))
	if err != nil {
		s.Logger.Error("Error parsing the raw diff from Github")
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
	s.Logger.Debug(fmt.Sprintf("Writting %d annotations to Github", len(comments)))
	checkRun, err := s.createCheckRun()
	if err != nil {
		s.Logger.Error("Error creating a Github check run")
		return err
	}
	if len(comments) == 0 {
		err := s.updateSuccessfulCheckRun(checkRun.GetID())
		if err != nil {
			s.Logger.Error("Error updating check run to success")
			return err
		}
		return nil
	}
	annotations := createAnnotations(comments)
	annotationLength := len(comments)
	summaryNumFindings := fmt.Sprintf(summaryString, annotationLength)
	// numIntermediateUpdateRequests contains the number of intermediate requests to be made prior to the final update request
	numIntermediateUpdateRequests := int(math.Ceil(float64(len(comments))/MaxAnnotationsPerRequest)) - 1
	for i := 0; i < numIntermediateUpdateRequests; i++ {
		startCommentIdx := i * MaxAnnotationsPerRequest
		endCommentIdx := min(startCommentIdx+MaxAnnotationsPerRequest, len(comments))
		opt := github.UpdateCheckRunOptions{
			Name: getCheckName(s.CheckRequest.Name),
			Output: &github.CheckRunOutput{
				Title:       github.String(getCheckName(s.CheckRequest.Name)),
				Annotations: annotations[startCommentIdx:endCommentIdx],
				Summary:     github.String(summaryNumFindings),
			},
		}
		_, _, err := s.Client.ChecksService().UpdateCheckRun(context.Background(),
			s.CheckRequest.Owner,
			s.CheckRequest.Repo,
			checkRun.GetID(),
			opt,
		)
		if err != nil {
			s.Logger.Warning("Unable to write 50 annotations to Github")
		}
	}
	remainingAnnotations := annotations[numIntermediateUpdateRequests*MaxAnnotationsPerRequest:]
	completedOpt := github.UpdateCheckRunOptions{
		Name:       getCheckName(s.CheckRequest.Name),
		Status:     &checkRunCompletedStatus,
		Conclusion: &checkRunConclusionFailure,
		Output: &github.CheckRunOutput{
			Title:       github.String(getCheckName(s.CheckRequest.Name)),
			Summary:     github.String(summaryNumFindings),
			Annotations: remainingAnnotations,
			Images: []*github.CheckRunImage{
				&github.CheckRunImage{
					Alt:      github.String(imageAlt),
					ImageURL: github.String(imageURL),
				},
			},
		},
	}
	_, _, err = s.Client.ChecksService().UpdateCheckRun(context.Background(),
		s.CheckRequest.Owner,
		s.CheckRequest.Repo,
		checkRun.GetID(),
		completedOpt,
	)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("Unable to update check run to failed and submit %d annotations", len(remainingAnnotations)))
		return err
	}
	return nil
}

func (s *Service) updateSuccessfulCheckRun(checkRunID int64) error {
	annotationLength := 0
	successfulSummary := fmt.Sprintf(summaryString, annotationLength)
	opt := github.UpdateCheckRunOptions{
		Name:       getCheckName(s.CheckRequest.Name),
		Status:     &checkRunCompletedStatus,
		Conclusion: &checkRunConclusionSuccess,
		Output: &github.CheckRunOutput{
			Title:   github.String(getCheckName(s.CheckRequest.Name)),
			Summary: github.String(successfulSummary),
			Images: []*github.CheckRunImage{
				&github.CheckRunImage{
					Alt:      github.String(imageAlt),
					ImageURL: github.String(imageURL),
				},
			},
		},
	}
	_, _, err := s.Client.ChecksService().UpdateCheckRun(context.Background(),
		s.CheckRequest.Owner,
		s.CheckRequest.Repo,
		checkRunID,
		opt,
	)
	if err != nil {
		return fmt.Errorf("failed to create check: %v", err)
	}
	return nil
}

// createCheckRun creates a new check run
func (s *Service) createCheckRun() (*github.CheckRun, error) {
	opt := github.CreateCheckRunOptions{
		Name:    getCheckName(s.CheckRequest.Name),
		HeadSHA: s.CheckRequest.SHA,
		Status:  &checkRunInProgressStatus,
	}
	checkRun, _, err := s.Client.ChecksService().CreateCheckRun(
		context.Background(),
		s.CheckRequest.Owner,
		s.CheckRequest.Repo,
		opt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create check: %v", err)
	}
	return checkRun, nil
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
		Title:           &comment.Title,
		Message:         &comment.Body,
		AnnotationLevel: &annotationLevelFailure,
	}
}

func getCheckName(name string) string {
	if name != "" {
		return name
	}
	return "Nightfall DLP"
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

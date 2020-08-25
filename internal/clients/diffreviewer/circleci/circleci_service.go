package circleci

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v31/github"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer/diffutils"
	gc "github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer/github"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/gitdiff"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/logger"
	circlelogger "github.com/nightfallai/nightfall_code_scanner/internal/clients/logger/circle_logger"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/nightfall"
	"github.com/nightfallai/nightfall_code_scanner/internal/interfaces/gitdiffintf"
	"github.com/nightfallai/nightfall_code_scanner/internal/interfaces/githubintf"
	"github.com/nightfallai/nightfall_code_scanner/internal/nightfallconfig"
)

const (
	WorkspacePathEnvVar          = "GITHUB_WORKSPACE"
	NightfallAPIKeyEnvVar        = "NIGHTFALL_API_KEY"
	CircleRepoNameEnvVar         = "CIRCLE_PROJECT_REPONAME"
	CircleOwnerNameEnvVar        = "CIRCLE_PROJECT_USERNAME"
	CircleCurrentCommitShaEnvVar = "CIRCLE_SHA1"
	CircleBeforeCommitEnvVar     = "EVENT_BEFORE"
	CirclePullRequestUrlEnvVar   = "CIRCLE_PULL_REQUEST"

	// right side is reserved for additions and unchanged lines
	// https://developer.github.com/v3/pulls/comments/#create-a-review-comment-for-a-pull-request
	GithubCommentRightSide = "RIGHT"
)

// Service contains the github client that makes Github api calls
type Service struct {
	GithubClient githubintf.GithubClient
	Logger       logger.Logger
	GitDiff      gitdiffintf.GitDiff
	PrDetails    prDetails
}

type prDetails struct {
	CommitSha string
	Owner     string
	Repo      string
	PrNumber  int
}

// NewCircleCiService creates a new CircleCi service
func NewCircleCiService(token string) diffreviewer.DiffReviewer {
	return &Service{
		GithubClient: gc.NewAuthenticatedClient(token),
		Logger:       circlelogger.NewDefaultCircleLogger(),
	}
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
		return nil, errors.New("missing env var for workspace path")
	}
	commitSha, ok := os.LookupEnv(CircleCurrentCommitShaEnvVar)
	if !ok {
		s.Logger.Error(fmt.Sprintf("Environment variable %s cannot be found", CircleCurrentCommitShaEnvVar))
		return nil, errors.New("missing env var for commit sha")
	}
	beforeCommitSha, ok := os.LookupEnv(CircleBeforeCommitEnvVar)
	if !ok {
		s.Logger.Error(fmt.Sprintf("Environment variable %s cannot be found", CircleBeforeCommitEnvVar))
		return nil, errors.New("missing env var for prev commit sha")
	}
	s.GitDiff = &gitdiff.GitDiff{
		WorkDir:    workspacePath,
		BaseBranch: "master", //TODO: look into how to get this instead of hardcoding
		BaseSHA:    beforeCommitSha,
		Head:       commitSha,
	}
	owner, ok := os.LookupEnv(CircleOwnerNameEnvVar)
	if !ok {
		s.Logger.Error(fmt.Sprintf("Environment variable %s cannot be found", CircleOwnerNameEnvVar))
		return nil, errors.New("missing env var for repo owner")
	}
	repo, ok := os.LookupEnv(CircleRepoNameEnvVar)
	if !ok {
		s.Logger.Error(fmt.Sprintf("Environment variable %s cannot be found", CircleRepoNameEnvVar))
		return nil, errors.New("missing env var for repository name")
	}
	prNumberUrl, ok := os.LookupEnv(CirclePullRequestUrlEnvVar)
	if !ok {
		s.Logger.Error(fmt.Sprintf("Environment variable %s cannot be found", CirclePullRequestUrlEnvVar))
		return nil, errors.New("missing env var for pull request url")
	}
	prNumber, err := strconv.Atoi(prNumberUrl[strings.LastIndex(prNumberUrl, "/")+1:])
	if err != nil {
		s.Logger.Error(fmt.Sprintf("Environment variable %s has an invalid format: %s", CirclePullRequestUrlEnvVar, prNumberUrl))
		return nil, errors.New("invalid format of pull request url env var")
	}
	s.PrDetails = prDetails{
		CommitSha: commitSha,
		Owner:     owner,
		Repo:      repo,
		PrNumber:  prNumber,
	}
	nightfallConfig, err := nightfallconfig.GetNightfallConfigFile(workspacePath, nightfallConfigFileName)
	if err != nil {
		s.Logger.Error("Error getting Nightfall config file. Ensure you have a Nightfall config file located in the root of your repository at .nightfalldlp/config.json with at least one Detector enabled")
		return nil, err
	}
	nightfallAPIKey, ok := os.LookupEnv(NightfallAPIKeyEnvVar)
	if !ok {
		s.Logger.Error(fmt.Sprintf("Error getting Nightfall API key. Ensure you have %s set in the Github secrets of the repo", NightfallAPIKeyEnvVar))
		return nil, errors.New("missing env var for nightfall api key")
	}

	var maxNumberRoutines int
	if nightfallConfig.MaxNumberRoutines < nightfall.MaxConcurrentRoutinesCap {
		maxNumberRoutines = nightfallConfig.MaxNumberRoutines
	} else {
		maxNumberRoutines = nightfall.MaxConcurrentRoutinesCap
	}
	return &nightfallconfig.Config{
		NightfallAPIKey:            nightfallAPIKey,
		NightfallDetectors:         nightfallConfig.Detectors,
		NightfallMaxNumberRoutines: maxNumberRoutines,
		TokenExclusionList:         nightfallConfig.TokenExclusionList,
		FileInclusionList:          nightfallConfig.FileInclusionList,
		FileExclusionList:          nightfallConfig.FileExclusionList,
	}, nil
}

// GetDiff retrieves the file diff from the requested pull request
func (s *Service) GetDiff() ([]*diffreviewer.FileDiff, error) {
	s.Logger.Debug("Getting diff from Github")
	content, err := s.GitDiff.GetDiff()
	if err != nil {
		s.Logger.Error(fmt.Sprintf("Error getting the raw diff from Github: %v", err))
		return nil, err
	}

	fileDiffs, err := diffutils.ParseMultiFile(strings.NewReader(content))
	if err != nil {
		s.Logger.Error("Error parsing the raw diff from Github")
		return nil, err
	}
	fileDiffs = diffutils.FilterFileDiffs(fileDiffs)
	return fileDiffs, nil
}

// WriteComments posts the findings as annotations to the github check
func (s *Service) WriteComments(comments []*diffreviewer.Comment) error {
	if len(comments) == 0 {
		return nil
	}
	githubComments := s.createGithubComments(comments)
	for _, c := range githubComments {
		_, _, err := s.GithubClient.PullRequestsService().CreateComment(
			context.Background(),
			s.PrDetails.Owner,
			s.PrDetails.Repo,
			s.PrDetails.PrNumber,
			c,
		)
		if err != nil {
			s.Logger.Error(fmt.Sprintf("Error writing comment to pull request: %s", err.Error()))
		}
	}
	// returning error to fail circleCI step
	return errors.New("potentially sensitive items found")
}

func (s *Service) createGithubComments(comments []*diffreviewer.Comment) []*github.PullRequestComment {
	githubComments := make([]*github.PullRequestComment, len(comments))
	for i, comment := range comments {
		githubComments[i] = &github.PullRequestComment{
			CommitID: &s.PrDetails.CommitSha,
			Body:     &comment.Body,
			Path:     &comment.FilePath,
			Line:     &comment.LineNumber,
			Side:     github.String(GithubCommentRightSide),
		}
	}
	return githubComments
}

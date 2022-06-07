package circleci

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v33/github"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer/diffutils"
	gc "github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer/github"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/gitdiff"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/logger"
	circlelogger "github.com/nightfallai/nightfall_code_scanner/internal/clients/logger/circle_logger"
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
	CircleBranchEnvVar           = "CIRCLE_BRANCH" // branch that triggered the workflow

	GithubBaseBranchEnvVar = "GITHUB_BASE_BRANCH" // optional user input variable if base branch is not master
	DefaultBaseBranchName  = "master"             // diff against base branch if workflow triggered by PR

	// right side is reserved for additions and unchanged lines
	// https://developer.github.com/v3/pulls/comments/#create-a-review-comment-for-a-pull-request
	GithubCommentRightSide = "RIGHT"
)

var errSensitiveItemsFound = errors.New("potentially sensitive items found")

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
	PrNumber  *int
}

// NewCircleCiService creates a new CircleCi service
func NewCircleCiService() diffreviewer.DiffReviewer {
	return &Service{
		Logger: circlelogger.NewDefaultCircleLogger(),
	}
}

// NewCircleCiServiceWithGithubComments creates a new CircleCi service with an authenticated Github client
func NewCircleCiServiceWithGithubComments(token, baseUrl string) diffreviewer.DiffReviewer {
	return &Service{
		GithubClient: gc.NewAuthenticatedClient(token, baseUrl),
		Logger:       circlelogger.NewDefaultCircleLogger(),
	}
}

// GetLogger gets the github service logger
func (s *Service) GetLogger() logger.Logger {
	return s.Logger
}

// LoadConfig gets all config values from files or environment and creates a config
func (s *Service) LoadConfig(nightfallConfigFileName string) (*nightfallconfig.Config, error) {
	s.Logger.Info("Loading configuration")
	workspacePath, ok := os.LookupEnv(WorkspacePathEnvVar)
	if !ok || workspacePath == "" {
		s.Logger.Error(fmt.Sprintf("Environment variable %s cannot be found", WorkspacePathEnvVar))
		return nil, errors.New("missing env var for workspace path")
	}
	beforeCommitSha, _ := os.LookupEnv(CircleBeforeCommitEnvVar)
	baseBranch, err := s.getBaseBranch()
	if err != nil {
		return nil, err
	}
	prDetails, err := s.getPrDetails()
	if err != nil {
		return nil, err
	}
	s.PrDetails = *prDetails
	s.GitDiff = &gitdiff.GitDiff{
		WorkDir:    workspacePath,
		BaseBranch: baseBranch,
		BaseSHA:    beforeCommitSha,
		Head:       s.PrDetails.CommitSha,
	}
	nightfallConfig, err := nightfallconfig.GetNightfallConfigFile(workspacePath, nightfallConfigFileName, s.Logger)
	if err != nil {
		s.Logger.Error("Error getting Nightfall config file. " +
			"Ensure you have a Nightfall config file located in the root of your repository at .nightfalldlp/config.json " +
			"with either a Condition Set UUID or at least one Condition enabled")
		return nil, err
	}
	nightfallAPIKey, ok := os.LookupEnv(NightfallAPIKeyEnvVar)
	if !ok || nightfallAPIKey == "" {
		s.Logger.Error(fmt.Sprintf("Error getting Nightfall API key. Ensure you have %s set in the Github secrets of the repo", NightfallAPIKeyEnvVar))
		return nil, errors.New("missing env var for nightfall api key")
	}
	return &nightfallconfig.Config{
		NightfallAPIKey:             nightfallAPIKey,
		NightfallDetectionRuleUUIDs: nightfallConfig.DetectionRuleUUIDs,
		NightfallDetectionRules:     nightfallConfig.DetectionRules,
		NightfallMaxNumberRoutines:  nightfallConfig.MaxNumberRoutines,
		TokenExclusionList:          nightfallConfig.TokenExclusionList,
		FileInclusionList:           nightfallConfig.FileInclusionList,
		FileExclusionList:           nightfallConfig.FileExclusionList,
		DefaultRedactionConfig:      nightfallConfig.DefaultRedactionConfig,
	}, nil
}

func (s *Service) getBaseBranch() (string, error) {
	var baseBranch string
	workflowBranch, ok := os.LookupEnv(CircleBranchEnvVar)
	if !ok {
		s.Logger.Error(fmt.Sprintf("Environment variable %s cannot be found", CircleBranchEnvVar))
		return "", errors.New("missing env var for branch name")
	}
	inputBaseBranch, ok := os.LookupEnv(GithubBaseBranchEnvVar)
	if ok && inputBaseBranch != "" {
		// don't set base branch if branch that triggered the workflow is the inputBaseBranch
		if inputBaseBranch != workflowBranch {
			baseBranch = inputBaseBranch
		}
	} else {
		// if master branch triggered the workflow, do not set the baseBranch
		if workflowBranch != DefaultBaseBranchName {
			baseBranch = DefaultBaseBranchName
		}
	}
	return baseBranch, nil
}

func (s *Service) getPrDetails() (*prDetails, error) {
	commitSha, ok := os.LookupEnv(CircleCurrentCommitShaEnvVar)
	if !ok || commitSha == "" {
		s.Logger.Error(fmt.Sprintf("Environment variable %s cannot be found", CircleCurrentCommitShaEnvVar))
		return nil, errors.New("missing env var for commit sha")
	}
	owner, ok := os.LookupEnv(CircleOwnerNameEnvVar)
	if !ok || owner == "" {
		s.Logger.Error(fmt.Sprintf("Environment variable %s cannot be found", CircleOwnerNameEnvVar))
		return nil, errors.New("missing env var for repo owner")
	}
	repo, ok := os.LookupEnv(CircleRepoNameEnvVar)
	if !ok || repo == "" {
		s.Logger.Error(fmt.Sprintf("Environment variable %s cannot be found", CircleRepoNameEnvVar))
		return nil, errors.New("missing env var for repository name")
	}
	prNumberUrl, ok := os.LookupEnv(CirclePullRequestUrlEnvVar)
	var prNumber *int
	if ok && prNumberUrl != "" {
		prNum, err := strconv.Atoi(prNumberUrl[strings.LastIndex(prNumberUrl, "/")+1:])
		if err != nil {
			s.Logger.Error(fmt.Sprintf("Environment variable %s has an invalid format: %s", CirclePullRequestUrlEnvVar, prNumberUrl))
			return nil, errors.New("invalid format of pull request url env var")
		}
		prNumber = &prNum
	}
	return &prDetails{
		CommitSha: commitSha,
		Owner:     owner,
		Repo:      repo,
		PrNumber:  prNumber,
	}, nil
}

// GetDiff retrieves the file diff from the requested pull request
func (s *Service) GetDiff() ([]*diffreviewer.FileDiff, error) {
	s.Logger.Info("Getting diff from Github")
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
func (s *Service) WriteComments(comments []*diffreviewer.Comment, level string) error {
	if len(comments) == 0 {
		s.Logger.Info("no sensitive items found")
		return nil
	}
	s.logCommentsToCircle(comments, level)
	if s.GithubClient == nil {
		return errSensitiveItemsFound
	}
	if s.PrDetails.PrNumber != nil {
		existingComments, _, err := s.GithubClient.PullRequestsService().ListComments(
			context.Background(),
			s.PrDetails.Owner,
			s.PrDetails.Repo,
			*s.PrDetails.PrNumber,
			&github.PullRequestListCommentsOptions{},
		)
		if err != nil {
			s.Logger.Error(fmt.Sprintf("Error listing existing pull request comments: %s", err.Error()))
		}
		githubComments := s.createGithubPullRequestComments(comments, level)
		filteredGithubComments := filterExistingComments(githubComments, existingComments)
		for _, c := range filteredGithubComments {
			_, _, err := s.GithubClient.PullRequestsService().CreateComment(
				context.Background(),
				s.PrDetails.Owner,
				s.PrDetails.Repo,
				*s.PrDetails.PrNumber,
				c,
			)
			if err != nil {
				s.Logger.Error(fmt.Sprintf("Error writing comment to pull request: %s", err.Error()))
			}
		}
	} else {
		githubComments := s.createGithubRepositoryComments(comments, level)
		for _, c := range githubComments {
			_, _, err := s.GithubClient.RepositoriesService().CreateComment(
				context.Background(),
				s.PrDetails.Owner,
				s.PrDetails.Repo,
				s.PrDetails.CommitSha,
				c,
			)
			if err != nil {
				s.Logger.Error(fmt.Sprintf("Error writing comment to commit: %s", err.Error()))
			}
		}
	}
	// returning error to fail circleCI step
	return errSensitiveItemsFound
}

func (s *Service) logCommentsToCircle(comments []*diffreviewer.Comment, level string) {
	for _, comment := range comments {
		s.Logger.Error(fmt.Sprintf(
			"%s: %s at %s on line %d",
			level,
			comment.Body,
			comment.FilePath,
			comment.LineNumber,
		))
	}
}

func (s *Service) createGithubPullRequestComments(comments []*diffreviewer.Comment, level string) []*github.PullRequestComment {
	githubComments := make([]*github.PullRequestComment, len(comments))
	for i, comment := range comments {
		body := fmt.Sprintf("%s: %s", level, comment.Body)
		githubComments[i] = &github.PullRequestComment{
			CommitID: &s.PrDetails.CommitSha,
			Body:     &body,
			Path:     &comment.FilePath,
			Line:     &comment.LineNumber,
			Side:     github.String(GithubCommentRightSide),
		}
	}
	return githubComments
}

func (s *Service) createGithubRepositoryComments(comments []*diffreviewer.Comment, level string) []*github.RepositoryComment {
	githubComments := make([]*github.RepositoryComment, len(comments))
	for i, comment := range comments {
		body := fmt.Sprintf("%s: %s", level, comment.Body)
		githubComments[i] = &github.RepositoryComment{
			CommitID: &s.PrDetails.CommitSha,
			Body:     &body,
			Path:     &comment.FilePath,
			Position: &comment.LineNumber,
		}
	}
	return githubComments
}

type prComment struct {
	Body string
	Path string
	Line int
}

func filterExistingComments(comments []*github.PullRequestComment, existingComments []*github.PullRequestComment) []*github.PullRequestComment {
	existingCommentsMap := make(map[prComment]bool, len(existingComments))
	for _, ec := range existingComments {
		comment := prComment{
			Body: ec.GetBody(),
			Path: ec.GetPath(),
			Line: ec.GetLine(),
		}
		existingCommentsMap[comment] = true
	}
	filteredComments := make([]*github.PullRequestComment, 0, len(comments))
	for _, c := range comments {
		comment := prComment{
			Body: c.GetBody(),
			Path: c.GetPath(),
			Line: c.GetLine(),
		}
		if _, ok := existingCommentsMap[comment]; !ok {
			filteredComments = append(filteredComments, c)
		}
	}
	return filteredComments
}

package circleci

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer/diffutils"

	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/gitdiff"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/logger"
	circlelogger "github.com/nightfallai/nightfall_code_scanner/internal/clients/logger/circle_logger"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/nightfall"
	"github.com/nightfallai/nightfall_code_scanner/internal/interfaces/gitdiffintf"
	"github.com/nightfallai/nightfall_code_scanner/internal/nightfallconfig"
)

const (
	WorkspacePathEnvVar        = "GITHUB_WORKSPACE"
	NightfallAPIKeyEnvVar      = "NIGHTFALL_API_KEY"
	CircleRepoNameEnvVar       = "CIRCLE_PROJECT_REPONAME"
	CircleOwnerNameEnvVar      = "CIRCLE_PROJECT_USERNAME"
	CircleCommitShaEnvVar      = "CIRCLE_SHA1"
	CircleBeforeCommitEnvVar   = "EVENT_BEFORE"
	CirclePullRequestUrlEnvVar = "CIRCLE_PULL_REQUEST"
)

// Service contains the github client that makes Github api calls
type Service struct {
	Logger  logger.Logger
	GitDiff gitdiffintf.GitDiff
}

// NewCircleCiService creates a new CircleCi service
func NewCircleCiService() diffreviewer.DiffReviewer {
	return &Service{
		Logger: circlelogger.NewDefaultCircleLogger(),
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
		return nil, errors.New("Missing env var for workspace path")
	}
	commitSha, ok := os.LookupEnv(CircleCommitShaEnvVar)
	if !ok {
		s.Logger.Error(fmt.Sprintf("Environment variable %s cannot be found", CircleCommitShaEnvVar))
		return nil, errors.New("Missing env var for commit sha")
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
	nightfallConfig, err := nightfallconfig.GetNightfallConfigFile(workspacePath, nightfallConfigFileName)
	if err != nil {
		s.Logger.Error("Error getting Nightfall config file. Ensure you have a Nightfall config file located in the root of your repository at .nightfalldlp/config.json with at least one Detector enabled")
		return nil, err
	}
	nightfallAPIKey, ok := os.LookupEnv(NightfallAPIKeyEnvVar)
	if !ok {
		s.Logger.Error(fmt.Sprintf("Error getting Nightfall API key. Ensure you have %s set in the Github secrets of the repo", NightfallAPIKeyEnvVar))
		return nil, errors.New("Missing env var for nightfall api key")
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
	//TODO: implement
	return nil
}

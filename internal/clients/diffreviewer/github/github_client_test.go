package github_test

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/suite"
	nightfallAPI "github.com/watchtowerai/nightfall_api/generated"
	"github.com/watchtowerai/nightfall_dlp/internal/clients/diffreviewer/github"
	"github.com/watchtowerai/nightfall_dlp/internal/nightfallconfig"
)

const testFileName = "nightfall_test_config.json"

var envVars = []string{
	github.WorkspacePathEnvVar,
	github.RepoEnvVar,
	github.CommitShaEnvVar,
	github.NightfallAPIKeyEnvVar,
}

type githubTestSuite struct {
	suite.Suite
}

func TestGithub(t *testing.T) {
	suite.Run(t, new(githubTestSuite))
}

func (g *githubTestSuite) AfterTest(suiteName, testName string) {
	for _, e := range envVars {
		err := os.Unsetenv(e)
		g.NoErrorf(err, "Error unsetting var %s", e)
	}
}

func (g *githubTestSuite) TestLoadConfig() {
	apiKey := "api-key"
	sha := "1234"
	owner := "nightfallai"
	repo := "testRepo"
	repoFullName := fmt.Sprintf("%s/%s", owner, repo)
	workspace, err := os.Getwd()
	g.NoError(err, "Error getting workspace")
	workspacePath := path.Join(workspace, "../../../../test/data")
	os.Setenv(github.WorkspacePathEnvVar, workspacePath)
	os.Setenv(github.RepoEnvVar, repoFullName)
	os.Setenv(github.CommitShaEnvVar, sha)
	os.Setenv(github.NightfallAPIKeyEnvVar, apiKey)

	expectedNightfallConfig := &nightfallconfig.Config{
		NightfallAPIKey: apiKey,
		NightfallDetectors: nightfallconfig.DetectorConfig{
			nightfallAPI.CREDIT_CARD_NUMBER: nightfallAPI.POSSIBLE,
			nightfallAPI.PHONE_NUMBER:       nightfallAPI.LIKELY,
		},
	}
	expectedGithubCheckRequest := github.CheckRequest{
		Owner: owner,
		Repo:  repo,
		SHA:   sha,
	}

	diffReviewer := github.NewGithubClient(&http.Client{})
	nightfallConfig, err := diffReviewer.LoadConfig(testFileName)
	g.NoError(err, "Error in LoadConfig")
	gh, ok := diffReviewer.(*github.Client)
	g.Equal(true, ok, "Error casting to github.Client")
	g.Equal(expectedNightfallConfig, nightfallConfig, "Incorrect nightfall config")
	g.Equal(expectedGithubCheckRequest, gh.CheckRequest, "Incorrect nightfall config")
}

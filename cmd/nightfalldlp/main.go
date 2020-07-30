package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/nightfallai/jenkins_test/internal/clients/flag"

	"github.com/nightfallai/jenkins_test/internal/clients/diffreviewer"
	"github.com/nightfallai/jenkins_test/internal/clients/diffreviewer/github"
	"github.com/nightfallai/jenkins_test/internal/clients/nightfall"
)

const (
	nightfallConfigFileName = ".nightfalldlp/config.json"
	githubActionsEnvVar     = "GITHUB_ACTIONS"
	githubTokenEnvVar       = "NIGHTFALL_GITHUB_TOKEN"
)

// main starts the service process.
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "nightfalldlp: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	flageValues := flag.Parse()

	diffReviewClient, err := CreateDiffReviewerClient(flageValues)
	if err != nil {
		return err
	}

	nightfallConfig, err := diffReviewClient.LoadConfig(nightfallConfigFileName)
	if err != nil {
		return err
	}
	nightfallClient := nightfall.NewClient(*nightfallConfig)

	fileDiffs, err := diffReviewClient.GetDiff()
	if err != nil {
		return err
	}

	comments, err := nightfallClient.ReviewDiff(ctx, diffReviewClient.GetLogger(), fileDiffs)
	if err != nil {
		return err
	}

	return diffReviewClient.WriteComments(comments)
}

// usingGithubAction determine if nightfalldlp is being run by
// Github Actions
func usingGithubAction() bool {
	val, ok := os.LookupEnv(githubActionsEnvVar)
	if !ok {
		return false
	}
	return val == "true"
}

// CreateDiffReviewerClient determines the current environment that is running nightfalldlp
// and returns the corresponding DiffReviewer client
func CreateDiffReviewerClient(flagValues *flag.Values) (diffreviewer.DiffReviewer, error) {
	switch {
	case usingGithubAction():
		githubToken, ok := os.LookupEnv(githubTokenEnvVar)
		if !ok {
			return nil, errors.New("missing github token in env")
		}
		return github.NewAuthenticatedGithubService(githubToken), nil
	default:
		return nil, errors.New("current environment unknown")
	}
}

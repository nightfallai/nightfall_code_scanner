package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer/circleci"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer/github"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/flag"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/nightfall"
)

const (
	nightfallConfigFileName = ".nightfalldlp/config.json"
	githubActionsEnvVar     = "GITHUB_ACTIONS"
	githubTokenEnvVar       = "GITHUB_TOKEN"
	circleCiEnvVar          = "CIRCLECI"
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
	_, done := flag.Parse(os.Args[1:])
	if done {
		return nil
	}

	diffReviewClient, err := CreateDiffReviewerClient()
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

// usingGithubAction determine if nightfalldlp is being triggered by CircleCi
func usingCircleCi() bool {
	val, ok := os.LookupEnv(circleCiEnvVar)
	if !ok {
		return false
	}
	return val == "true"
}

// CreateDiffReviewerClient determines the current environment that is running nightfalldlp
// and returns the corresponding DiffReviewer client
func CreateDiffReviewerClient() (diffreviewer.DiffReviewer, error) {
	switch {
	case usingGithubAction():
		githubToken, ok := os.LookupEnv(githubTokenEnvVar)
		if !ok {
			return nil, fmt.Errorf("could not find required %s environment variable", githubTokenEnvVar)
		}
		return github.NewAuthenticatedGithubService(githubToken), nil
	case usingCircleCi():
		githubToken, ok := os.LookupEnv(githubTokenEnvVar)
		if !ok {
			return circleci.NewUnauthenticatedCircleCiService(), nil
		}
		return circleci.NewAuthenticatedCircleCiService(githubToken), nil
	default:
		return nil, errors.New("current environment unknown")
	}
}

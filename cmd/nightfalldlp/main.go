package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/watchtowerai/nightfall_dlp/internal/clients/diffreviewer"
	"github.com/watchtowerai/nightfall_dlp/internal/clients/diffreviewer/github"
)

const (
	githubActionsEnvVar = "GITHUB_ACTIONS"
)

// main starts the service process.
func main() {
	_, err := CreateDiffReviewerClient(&http.Client{})
	if err != nil {
		fmt.Printf("Error Getting Client %v", err)
	}
	fmt.Println("Running NightfallDLP Action")
}

// usingGithubAction determine if nightfalldpl is being run by
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
func CreateDiffReviewerClient(httpClient *http.Client) (diffreviewer.DiffReviewer, error) {
	switch {
	case usingGithubAction():
		return github.NewGithubClient(httpClient), nil
	default:
		return nil, errors.New("Current environment unknown")
	}
}

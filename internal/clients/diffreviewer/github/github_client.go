package github

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v31/github"
	"github.com/watchtowerai/nightfall_dlp/internal/clients/diffreviewer"
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
type Client struct {
	*github.Client
	CheckRequest CheckRequest
	CheckRun     *github.CheckRun
}

// NewGithubClient create a new github client
func NewGithubClient(httpClient *http.Client) diffreviewer.DiffReviewer {
	client := github.NewClient(httpClient)
	return &Client{Client: client}
}

// LoadConfig gets all config values from files or environment and creates a config
func (c *Client) LoadConfig(nightfallConfigFileName string) (*nightfallconfig.Config, error) {
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
	c.CheckRequest = CheckRequest{
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
func (c *Client) GetDiff() ([]*diffreviewer.FileDiff, error) {
	//TODO implement
	return nil, nil
}

// WriteComments posts the findings as annotations to the github check
func (c *Client) WriteComments(
	comments []*diffreviewer.Comment,
) error {
	//TODO implement
	return nil
}

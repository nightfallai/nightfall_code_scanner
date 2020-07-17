package github

import (
	"context"
	"net/http"

	"github.com/google/go-github/v31/github"
	"github.com/watchtowerai/nightfall_dlp/internal/interfaces"
)

type Client struct {
	*github.Client
}

func NewClient(httpClientInterface interfaces.HTTPClient) *Client {
	httpClient := httpClientInterface.(*http.Client)
	githubClient := github.NewClient(httpClient)
	return &Client{githubClient}
}

// GetRaw gets a single pull request in raw (diff or patch) format.
func (c *Client) GetRaw(ctx context.Context, owner string, repo string, number int, opts github.RawOptions) (string, *github.Response, error) {
	return c.Client.PullRequests.GetRaw(ctx, owner, repo, number, opts)
}

// CreateCheckRun creates a new check run for a specific commit in a repository. Your GitHub App must have the checks:write permission to create check runs.
func (c *Client) CreateCheckRun(ctx context.Context, owner, repo string, opts github.CreateCheckRunOptions) (*github.CheckRun, *github.Response, error) {
	return c.Client.Checks.CreateCheckRun(ctx, owner, repo, opts)
}

// UpdateCheckRun updates a check run for a specific commit in a repository. Your GitHub App must have the checks:write permission to edit check runs.
func (c *Client) UpdateCheckRun(ctx context.Context, owner, repo string, checkRunID int64, opts github.UpdateCheckRunOptions) (*github.CheckRun, *github.Response, error) {
	return c.Client.Checks.UpdateCheckRun(ctx, owner, repo, checkRunID, opts)
}

package github

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v31/github"
	"github.com/watchtowerai/nightfall_dlp/internal/interfaces"
)

const mediaTypeV3Diff = "application/vnd.github.v3.diff"

// Client is a wrapper around github.Client
type Client struct {
	*github.Client
}

// NewClient generates a new github client
func NewClient(httpClientInterface interfaces.HTTPClient) *Client {
	httpClient := httpClientInterface.(*http.Client)
	githubClient := github.NewClient(httpClient)
	return &Client{githubClient}
}

// GetRaw gets a single pull request in raw (diff or patch) format.
func (c *Client) GetRaw(ctx context.Context, owner string, repo string, number int, opts github.RawOptions) (string, *github.Response, error) {
	return c.Client.PullRequests.GetRaw(ctx, owner, repo, number, opts)
}

// GetRawBySha gets the diff based on base and head commit
func (c *Client) GetRawBySha(ctx context.Context, owner string, repo string, sha string, head string) (string, *github.Response, error) {
	u := fmt.Sprintf("repos/%v/%v/compare/%v...%v", owner, repo, sha, head)
	req, err := c.Client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Accept", mediaTypeV3Diff)
	var buf bytes.Buffer
	resp, err := c.Client.Do(ctx, req, &buf)
	if err != nil {
		return "", nil, err
	}
	return buf.String(), resp, nil
}

// CreateCheckRun creates a new check run for a specific commit in a repository. Your GitHub App must have the checks:write permission to create check runs.
func (c *Client) CreateCheckRun(ctx context.Context, owner, repo string, opts github.CreateCheckRunOptions) (*github.CheckRun, *github.Response, error) {
	return c.Client.Checks.CreateCheckRun(ctx, owner, repo, opts)
}

// UpdateCheckRun updates a check run for a specific commit in a repository. Your GitHub App must have the checks:write permission to edit check runs.
func (c *Client) UpdateCheckRun(ctx context.Context, owner, repo string, checkRunID int64, opts github.UpdateCheckRunOptions) (*github.CheckRun, *github.Response, error) {
	return c.Client.Checks.UpdateCheckRun(ctx, owner, repo, checkRunID, opts)
}

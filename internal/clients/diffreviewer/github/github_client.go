package github

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v31/github"
	"github.com/watchtowerai/nightfall_dlp/internal/interfaces"
	"golang.org/x/oauth2"
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

func NewAuthenticatedClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	githubClient := github.NewClient(tc)
	return &Client{githubClient}
}

// GetRaw gets a single pull request in raw (diff or patch) format.
func (c *Client) GetRaw(
	ctx context.Context,
	owner string,
	repo string,
	number int,
	opts github.RawOptions,
) (string, *github.Response, error) {
	return c.Client.PullRequests.GetRaw(ctx, owner, repo, number, opts)
}

// GetRawBySha gets the diff based on the base and head commits (or branches)
// https://developer.github.com/v3/repos/commits/#compare-two-commits
func (c *Client) GetRawBySha(
	ctx context.Context,
	owner string,
	repo string,
	base string,
	headOrSha string,
) (string, *github.Response, error) {
	u := fmt.Sprintf("repos/%v/%v/compare/%v...%v", owner, repo, base, headOrSha)
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

// ListCheckRunsForRef lists all the check runs associated with the input ref (sha)
func (c *Client) ListCheckRunsForRef(ctx context.Context, owner, repo, ref string, opts *github.ListCheckRunsOptions) (*github.ListCheckRunsResults, *github.Response, error) {
	return c.Client.Checks.ListCheckRunsForRef(ctx, owner, repo, ref, opts)
}

// CreateCheckRun creates a new check run for a specific commit in a repository. Your GitHub App must have the checks:write permission to create check runs.
func (c *Client) CreateCheckRun(
	ctx context.Context,
	owner string,
	repo string,
	opts github.CreateCheckRunOptions,
) (*github.CheckRun, *github.Response, error) {
	return c.Client.Checks.CreateCheckRun(ctx, owner, repo, opts)
}

// UpdateCheckRun updates a check run for a specific commit in a repository. Your GitHub App must have the checks:write permission to edit check runs.
func (c *Client) UpdateCheckRun(
	ctx context.Context,
	owner string,
	repo string,
	checkRunID int64,
	opts github.UpdateCheckRunOptions,
) (*github.CheckRun, *github.Response, error) {
	return c.Client.Checks.UpdateCheckRun(ctx, owner, repo, checkRunID, opts)
}

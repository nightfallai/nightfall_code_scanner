package github

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v31/github"
	"github.com/watchtowerai/nightfall_dlp/internal/interfaces"
	"github.com/watchtowerai/nightfall_dlp/internal/interfaces/githubintf"
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

// NewAuthenticatedClient generates an authenticated github client
func NewAuthenticatedClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	githubClient := github.NewClient(tc)
	return &Client{githubClient}
}

// PullRequestService gets the github clients pull request service
func (c *Client) PullRequestService() githubintf.GithubPullRequest {
	return c.Client.PullRequests
}

// ChecksService gets the github clients checks service
func (c *Client) ChecksService() githubintf.GithubChecks {
	return c.Client.Checks
}

// Do completes a request to the github API
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*github.Response, error) {
	return c.Client.Do(ctx, req, v)
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
	resp, err := c.Do(ctx, req, &buf)
	if err != nil {
		return "", nil, err
	}
	return buf.String(), resp, nil
}

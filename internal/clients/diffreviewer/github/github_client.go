package github

import (
	"context"
	"net/url"
	"strings"

	"github.com/google/go-github/v31/github"
	"github.com/nightfallai/nightfall_code_scanner/internal/interfaces/githubintf"
	"golang.org/x/oauth2"
)

// Client is a wrapper around github.Client
type Client struct {
	*github.Client
}

// NewAuthenticatedClient generates an authenticated github client
func NewAuthenticatedClient(token string, baseUrl string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	githubClient := github.NewClient(tc)
	// for enterprise
	if baseUrl != "" {
		u, _ := url.Parse(baseUrl)
		if !strings.HasSuffix(u.Path, "/") {
			u.Path += "/"
		}
		if !strings.HasSuffix(u.Path, "/api/v3/") {
			u.Path += "api/v3/"
		}
		githubClient.BaseURL = u
	}
	return &Client{githubClient}
}

// ChecksService gets the github client's checks service
func (c *Client) ChecksService() githubintf.GithubChecks {
	return c.Client.Checks
}

// PullRequestsService gets the github client's pull requests service
func (c *Client) PullRequestsService() githubintf.GithubPullRequests {
	return c.Client.PullRequests
}

// RepositoriesService gets the github client's repositories service
func (c *Client) RepositoriesService() githubintf.GithubRepositories {
	return c.Client.Repositories
}

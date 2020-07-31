package github

import (
	"context"

	"github.com/google/go-github/v31/github"
	"github.com/nightfallai/jenkins_test/internal/interfaces/githubintf"
	"golang.org/x/oauth2"
)

// Client is a wrapper around github.Client
type Client struct {
	*github.Client
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

// ChecksService gets the github clients checks service
func (c *Client) ChecksService() githubintf.GithubChecks {
	return c.Client.Checks
}

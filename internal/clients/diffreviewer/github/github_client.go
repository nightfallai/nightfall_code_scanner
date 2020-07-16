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

func (c *Client) GetRaw(ctx context.Context, owner string, repo string, number int, opts github.RawOptions) (string, *github.Response, error) {
	return c.Client.PullRequests.GetRaw(ctx, owner, repo, number, opts)
}

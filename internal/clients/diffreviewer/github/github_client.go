package github

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-github/v33/github"
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

	var u *url.URL
	// for enterprise

	githubClient := github.NewClient(tc)
	if baseUrl != "" {
		u, _ = url.Parse(baseUrl)
		if !strings.HasSuffix(u.Path, "/") {
			u.Path += "/"
		}
		if !strings.HasSuffix(u.Path, "/api/v3/") {
			u.Path += "api/v3/"
		}
		githubClient.BaseURL = u
	}

	//
	logger := log.New(os.Stdout, "", 0)
	urlStr := fmt.Sprintf("%srepos/%v/%v/pulls/comments", u.String(), "nfdev456", "TestAction2")

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		fmt.Println("ERR Creating request")
	}
	req.Header.Set("Accept", "application/vnd.github.squirrel-girl-preview,application/vnd.github.comfort-fade-preview+json")
	resp, err := tc.Do(req)

	if err != nil {
		logger.Printf("Error making request: %s", err.Error())
	}
	logger.Printf("Response Status: %d", resp.StatusCode)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	logger.Printf("Response Body: %s", string(bodyBytes))
	if err != nil {
		logger.Printf("error reading from body: %s", err.Error())
	}
	//
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

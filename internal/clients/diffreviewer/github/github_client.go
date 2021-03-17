package github

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"moul.io/http2curl"

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
	//ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	//tc := oauth2.NewClient(ctx, ts)

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	//tc.Transport = customTransport
	tc := &http.Client{
		Transport: &oauth2.Transport{
			Base:   customTransport,
			Source: oauth2.ReuseTokenSource(nil, ts),
		},
	}

	/*
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client := &http.Client{Transport: customTransport}
	*/
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

	opt := github.CreateCheckRunOptions{
		Name:    getCheckName("TestAction"),
		HeadSHA: "d0c2aec77b2dd022dba20233c62b74eb63559032",
		Status:  &checkRunInProgressStatus,
	}

	var buf io.ReadWriter
	buf = &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(opt)
	if err != nil {
		logger.Printf("Error encoding opt request body: %s", err.Error())
	}
	owner := "alan20854"
	repo := "TestRepo2"
	//urlStr := fmt.Sprintf("%srepos/%v/%v/pulls/comments", u.String(), "alan20854", "TestRepo2")
	urlStr := fmt.Sprintf("%srepos/%v/%v/check-runs", u.String(), owner, repo)
	req, err := http.NewRequest(http.MethodPost, urlStr, buf)
	command, _ := http2curl.GetCurlCommand(req)
	logger.Printf("http2Curl output: %s", command.String())
	if err != nil {
		fmt.Println("ERR Creating request")
	}
	// for pulls/comments
	// req.Header.Set("Accept", "application/vnd.github.squirrel-girl-preview,application/vnd.github.comfort-fade-preview+json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
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

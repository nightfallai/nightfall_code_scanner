package githubintf

import (
	"context"
	"net/http"

	"github.com/google/go-github/v31/github"
)

//go:generate go run github.com/golang/mock/mockgen -destination=../../mocks/clients/githubclient_mock/githubclient_mock.go -source=../githubintf/github_client.go -package=githubclient_mock -mock_names=GithubClient=GithubClient

type GithubClient interface {
	PullRequestService() GithubPullRequest
	ChecksService() GithubChecks
	Do(ctx context.Context, req *http.Request, v interface{}) (*github.Response, error)
	GetRawBySha(ctx context.Context, owner string, repo string, base string, headOrSha string) (string, *github.Response, error)
}

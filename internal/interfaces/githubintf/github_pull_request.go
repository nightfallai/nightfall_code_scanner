package githubintf

import (
	"context"

	"github.com/google/go-github/v31/github"
)

//go:generate go run github.com/golang/mock/mockgen -destination=../../mocks/clients/githubpullrequest_mock/githubpullrequest_mock.go -source=../githubintf/github_pull_request.go -package=githubpullrequest_mock -mock_names=GithubPullRequest=GithubPullRequest

type GithubPullRequest interface {
	GetRaw(ctx context.Context, owner string, repo string, number int, opts github.RawOptions) (string, *github.Response, error)
}

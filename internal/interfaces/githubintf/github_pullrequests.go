package githubintf

import (
	"context"

	"github.com/google/go-github/v33/github"
)

//go:generate go run github.com/golang/mock/mockgen -destination=../../mocks/clients/githubpullrequests_mock/githubpullrequests_mock.go -source=../githubintf/github_pullrequests.go -package=githubpullrequests_mock -mock_names=GithubPullRequests=GithubPullRequests

type GithubPullRequests interface {
	CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.PullRequestComment) (*github.PullRequestComment, *github.Response, error)
	ListComments(ctx context.Context, owner string, repo string, number int, opts *github.PullRequestListCommentsOptions) ([]*github.PullRequestComment, *github.Response, error)
}

package githubintf

import (
	"context"

	"github.com/google/go-github/v31/github"
)

//go:generate go run github.com/golang/mock/mockgen -destination=../../mocks/clients/githubrepositories_mock/githubrepositories_mock.go -source=../githubintf/github_repositories.go -package=githubrepositories_mock -mock_names=GithubRepositories=GithubRepositories

type GithubRepositories interface {
	CreateComment(ctx context.Context, owner, repo, sha string, comment *github.RepositoryComment) (*github.RepositoryComment, *github.Response, error)
}

package interfaces

import (
	"context"

	"github.com/google/go-github/v31/github"
)

//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/clients/githubapi_mock/api_mock.go -source=../interfaces/github_api.go -package=githubapi_mock -mock_names=GithubAPI=GithubAPI

type GithubAPI interface {
	GetRaw(ctx context.Context, owner string, repo string, number int, opts github.RawOptions) (string, *github.Response, error)
}

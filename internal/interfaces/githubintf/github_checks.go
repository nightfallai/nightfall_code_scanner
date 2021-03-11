package githubintf

import (
	"context"

	"github.com/google/go-github/v33/github"
)

//go:generate go run github.com/golang/mock/mockgen -destination=../../mocks/clients/githubchecks_mock/githubchecks_mock.go -source=../githubintf/github_checks.go -package=githubchecks_mock -mock_names=GithubChecks=GithubChecks

type GithubChecks interface {
	CreateCheckRun(ctx context.Context, owner, repo string, opts github.CreateCheckRunOptions) (*github.CheckRun, *github.Response, error)
	UpdateCheckRun(ctx context.Context, owner, repo string, checkRunID int64, opts github.UpdateCheckRunOptions) (*github.CheckRun, *github.Response, error)
}

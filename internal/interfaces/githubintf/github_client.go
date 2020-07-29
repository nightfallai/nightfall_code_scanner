package githubintf

//go:generate go run github.com/golang/mock/mockgen -destination=../../mocks/clients/githubclient_mock/githubclient_mock.go -source=../githubintf/github_client.go -package=githubclient_mock -mock_names=GithubClient=GithubClient

type GithubClient interface {
	ChecksService() GithubChecks
}

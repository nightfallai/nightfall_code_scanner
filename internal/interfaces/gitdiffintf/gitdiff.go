package gitdiffintf

//go:generate go run github.com/golang/mock/mockgen -destination=../../mocks/clients/gitdiff_mock/gitdiff_mock.go -source=../gitdiffintf/gitdiff.go -package=gitdiff_mock -mock_names=GitDiff=GitDiff

type GitDiff interface {
	GetDiff() (string, error)
}

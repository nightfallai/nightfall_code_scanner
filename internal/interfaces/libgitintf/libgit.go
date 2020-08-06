package libgitintf

import libgit2 "github.com/libgit2/git2go/v30"

//go:generate go run github.com/golang/mock/mockgen -destination=../../mocks/clients/libgit_mock/libgit_mock.go -source=../libgitintf/libgit.go -package=libgit_mock -mock_names=Libgit=Libgit

type Libgit interface {
	Clone(repoURL, filePath string) (*libgit2.Repository, error)
	DiffRevToRev(repo *libgit2.Repository, baseRev, headRev string) (*libgit2.Diff, error)
	ConvertDiffToPatch(diff *libgit2.Diff) ([]byte, error)
}

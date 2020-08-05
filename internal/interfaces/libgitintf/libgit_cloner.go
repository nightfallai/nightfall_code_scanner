package libgitintf

import libgit2 "github.com/libgit2/git2go/v30"

//go:generate go run github.com/golang/mock/mockgen -destination=../../mocks/clients/libgitcloner_mock/libgitcloner_mock.go -source=../libgitintf/libgit_cloner.go -package=libgitcloner_mock -mock_names=LibgitCloner=LibgitCloner

type LibgitCloner interface {
	Clone(repoURL, filePath string) (*libgit2.Repository, error)
}

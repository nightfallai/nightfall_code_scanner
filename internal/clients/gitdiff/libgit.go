package gitdiff

import (
	libgit2 "github.com/libgit2/git2go/v30"
)

// Libgit is a wrapper around libgit2
type Libgit struct {
	accessToken string
}

// NewLibgit creates a new Libgit
func NewLibgit(accessToken string) *Libgit {
	return &Libgit{
		accessToken: accessToken,
	}
}

// Clone clones the repo to the desired file path
func (lg *Libgit) Clone(repoURL, filePath string) (*libgit2.Repository, error) {
	credCallback := func(url string, username_from_url string, allowed_types libgit2.CredType) (*libgit2.Cred, error) {
		return libgit2.NewCredUserpassPlaintext(lg.accessToken, "")
	}
	callbacks := libgit2.RemoteCallbacks{
		CredentialsCallback: credCallback,
	}
	cloneOps := &libgit2.CloneOptions{
		FetchOptions: &libgit2.FetchOptions{
			RemoteCallbacks: callbacks,
		},
	}
	repo, err := libgit2.Clone(repoURL, filePath, cloneOps)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// OpenRepo opens a repo which already exists in the file system
func (lg *Libgit) OpenRepo(filePath string) (*libgit2.Repository, error) {
	repo, err := libgit2.OpenRepository(filePath)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// DiffRevToRev gets the diff between two git revisions
// a revision can be a commit sha or a branch reference https://git-scm.com/docs/git-rev-parse.html#_specifying_revisions
func (lg *Libgit) DiffRevToRev(repo *libgit2.Repository, baseRev, headRev string) (*libgit2.Diff, error) {
	baseTree, err := getTreeForRev(repo, baseRev)
	if err != nil {
		return nil, err
	}
	headTree, err := getTreeForRev(repo, headRev)
	if err != nil {
		return nil, err
	}

	diff, err := repo.DiffTreeToTree(baseTree, headTree, nil)
	if err != nil {
		return nil, err
	}

	return diff, nil
}

// ConvertDiffToPatch converts the libgit2 diff to a patch string
func (lg *Libgit) ConvertDiffToPatch(diff *libgit2.Diff) ([]byte, error) {
	result, err := diff.ToBuf(libgit2.DiffFormatPatch)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func getTreeForRev(repo *libgit2.Repository, rev string) (*libgit2.Tree, error) {
	revObj, err := repo.RevparseSingle(rev)
	if err != nil {
		return nil, err
	}
	commit, err := revObj.AsCommit()
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	return tree, nil
}

package libgit

import libgit2 "github.com/libgit2/git2go/v30"

// Cloner clones repo
type Cloner struct {
	accessToken string
	repoURL     string
}

// NewCloner creates a new cloner
func NewCloner(accessToken, repoURL string) *Cloner {
	return &Cloner{
		accessToken: accessToken,
		repoURL:     repoURL,
	}
}

// Clone clones the repo to the desired file path
func (c *Cloner) Clone(filePath string) (*libgit2.Repository, error) {
	credCallback := func(url string, username_from_url string, allowed_types libgit2.CredType) (*libgit2.Cred, error) {
		return libgit2.NewCredUserpassPlaintext(c.accessToken, "")
	}
	callbacks := libgit2.RemoteCallbacks{
		CredentialsCallback: credCallback,
	}
	cloneOps := &libgit2.CloneOptions{
		FetchOptions: &libgit2.FetchOptions{
			RemoteCallbacks: callbacks,
		},
	}
	repo, err := libgit2.Clone(c.repoURL, filePath, cloneOps)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// OpenRepo opens a repo which already exists in the file system
func (c *Cloner) OpenRepo(filePath string) (*libgit2.Repository, error) {
	repo, err := libgit2.OpenRepository(filePath)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

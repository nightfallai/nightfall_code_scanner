package gitdiff

import (
	"bytes"

	"github.com/nightfallai/jenkins_test/internal/clients/diffreviewer"
	"github.com/nightfallai/jenkins_test/internal/interfaces/libgitintf"
)

// Client is a wrapper around the libgit interface
type Client struct {
	Libgit       libgitintf.Libgit
	RepoFilePath string
}

// DiffOptions options specifying the file diffs to be returned
type DiffOptions struct {
	Filter map[diffreviewer.LineType]bool
}

// NewClient creates a libgit client
func NewClient(accessToken, repoFilePath string) *Client {
	return &Client{
		Libgit:       NewLibgit(accessToken),
		RepoFilePath: repoFilePath,
	}
}

// GetDiff gets the diff from the base to the head on the given repo
func (c *Client) GetDiff(baseRev, headRev, repoURL string, diffOpts *DiffOptions) ([]*diffreviewer.FileDiff, error) {
	repo, err := c.Libgit.Clone(repoURL, c.RepoFilePath)
	if err != nil {
		return nil, err
	}

	diff, err := c.Libgit.DiffRevToRev(repo, baseRev, headRev)
	if err != nil {
		return nil, err
	}

	fileDiffBytes, err := c.Libgit.ConvertDiffToPatch(diff)
	if err != nil {
		return nil, err
	}

	fileDiffs, err := ParseMultiFile(bytes.NewReader(fileDiffBytes))
	if err != nil {
		return nil, err
	}
	finalFileDiffs := filterFileDiffs(fileDiffs, diffOpts)

	return finalFileDiffs, nil
}

func filterFileDiffs(fileDiffs []*diffreviewer.FileDiff, diffOpts *DiffOptions) []*diffreviewer.FileDiff {
	if len(fileDiffs) == 0 {
		return fileDiffs
	}
	filteredFileDiffs := []*diffreviewer.FileDiff{}
	for _, fileDiff := range fileDiffs {
		fileDiff.Hunks = filterHunks(fileDiff.Hunks, diffOpts)
		if len(fileDiff.Hunks) > 0 {
			filteredFileDiffs = append(filteredFileDiffs, fileDiff)
		}
	}
	return filteredFileDiffs
}

func filterHunks(hunks []*diffreviewer.Hunk, diffOpts *DiffOptions) []*diffreviewer.Hunk {
	filteredHunks := []*diffreviewer.Hunk{}
	for _, hunk := range hunks {
		hunk.Lines = filterLines(hunk.Lines, diffOpts)
		if len(hunk.Lines) > 0 {
			filteredHunks = append(filteredHunks, hunk)
		}
	}
	return filteredHunks
}

func filterLines(lines []*diffreviewer.Line, diffOpts *DiffOptions) []*diffreviewer.Line {
	filteredLines := []*diffreviewer.Line{}
	for _, line := range lines {
		if val, ok := diffOpts.Filter[line.Type]; ok && val {
			filteredLines = append(filteredLines, line)
		}
	}
	return filteredLines
}

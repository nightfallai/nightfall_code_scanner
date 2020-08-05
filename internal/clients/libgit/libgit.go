package libgit

import (
	"errors"

	libgit2 "github.com/libgit2/git2go/v30"
	"github.com/nightfallai/jenkins_test/internal/clients/diffreviewer"
	"github.com/nightfallai/jenkins_test/internal/interfaces/libgitintf"
)

// Client is a wrapper around libgit2
type Client struct {
	base   string
	head   string
	cloner libgitintf.LibgitCloner
}

type fileDiffHolder struct {
	fileDiff *diffreviewer.FileDiff
	hunks    []*diffreviewer.Hunk
}

// NewClient creates a libgit client
func NewClient(accessToken, base, head, repoURL string) *Client {
	return &Client{
		base:   base,
		head:   head,
		cloner: NewCloner(accessToken, repoURL),
	}
}

// GetDiff gets the diff from the base to the head on the given repo
func (c *Client) GetDiff() ([]*diffreviewer.FileDiff, error) {
	filePath := ""
	repo, err := c.cloner.Clone(filePath)
	if err != nil {
		return nil, err
	}

	baseTree, err := getTreeByHash(repo, c.base)
	if err != nil {
		return nil, err
	}
	headTree, err := getTreeByHash(repo, c.head)
	if err != nil {
		return nil, err
	}

	diff, err := repo.DiffTreeToTree(baseTree, headTree, nil)
	if err != nil {
		return nil, err
	}

	fileDiffs, err := convertDiffToFileDiffs(diff)
	if err != nil {
		return nil, err
	}

	return fileDiffs, nil
}

func convertDiffToFileDiffs(diff *libgit2.Diff) ([]*diffreviewer.FileDiff, error) {
	var lineCb libgit2.DiffForEachLineCallback
	var hunkCb libgit2.DiffForEachHunkCallback
	var fileCb libgit2.DiffForEachFileCallback

	fileDiffs := []*diffreviewer.FileDiff{}

	fileCb = func(delta libgit2.DiffDelta, _ float64) (libgit2.DiffForEachHunkCallback, error) {
		fileDiff := &diffreviewer.FileDiff{
			PathOld: delta.OldFile.Path,
			PathNew: delta.NewFile.Path,
			Hunks:   []*diffreviewer.Hunk{},
		}
		hunkCb = func(fileHunk libgit2.DiffHunk) (libgit2.DiffForEachLineCallback, error) {
			hunkTotalLines := fileHunk.NewLines + fileHunk.OldLines
			hunkLineNum := 0
			hunk := &diffreviewer.Hunk{
				StartLineOld:  fileHunk.OldStart,
				LineLengthOld: fileHunk.OldLines,
				StartLineNew:  fileHunk.NewStart,
				LineLengthNew: fileHunk.NewLines,
				Section:       fileHunk.Header,
				Lines:         []*diffreviewer.Line{},
			}
			lineCb = func(fileLine libgit2.DiffLine) error {
				// We only care about line addtions or line contexts
				if isDiffLineAdditionOrDiffLineDeleteOrDiffLineContext(fileLine) {
					return nil
				}
				hunkLineNum++
				fileType, err := convertFileLineOriginToLineType(fileLine.Origin)
				if err != nil {
					return err
				}
				line := &diffreviewer.Line{
					Type:     fileType,
					Content:  fileLine.Content,
					LnumDiff: fileLine.NewLineno,
					LnumOld:  fileLine.OldLineno,
					LnumNew:  fileLine.NewLineno,
				}
				hunk.Lines = append(hunk.Lines, line)
				// if we have appended all the lines to the hunk append hunk to file diff map
				if hunkLineNum == hunkTotalLines {
					fileDiff.Hunks = append(fileDiff.Hunks, hunk)
				}
				return nil
			}
			return lineCb, nil
		}
		return hunkCb, nil
	}

	err := diff.ForEach(fileCb, libgit2.DiffDetailLines)
	if err != nil {
		return nil, err
	}

	return fileDiffs, nil
}

func getTreeByHash(repo *libgit2.Repository, hash string) (*libgit2.Tree, error) {
	hashObj, err := repo.RevparseSingle(hash)
	if err != nil {
		return nil, err
	}
	commit, err := hashObj.AsCommit()
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func isDiffLineAdditionOrDiffLineDeleteOrDiffLineContext(fileLine libgit2.DiffLine) bool {
	return (fileLine.Origin != libgit2.DiffLineAddition) && (fileLine.Origin != libgit2.DiffLineDeletion) && (fileLine.Origin != libgit2.DiffLineContext)
}

func convertFileLineOriginToLineType(origin libgit2.DiffLineType) (diffreviewer.LineType, error) {
	switch origin {
	case libgit2.DiffLineAddition:
		return diffreviewer.LineAdded, nil
	case libgit2.DiffLineDeletion:
		return diffreviewer.LineDeleted, nil
	case libgit2.DiffLineContext:
		return diffreviewer.LineUnchanged, nil
	}
	return 0, errors.New("Unknown file line origin type")
}

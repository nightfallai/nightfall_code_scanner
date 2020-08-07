package gitdiff_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	libgit2 "github.com/libgit2/git2go/v30"
	"github.com/nightfallai/jenkins_test/internal/clients/diffreviewer"
	"github.com/nightfallai/jenkins_test/internal/clients/gitdiff"
	"github.com/nightfallai/jenkins_test/internal/mocks/clients/libgit_mock"
	"github.com/stretchr/testify/suite"
)

type gitdiffTestSuite struct {
	suite.Suite
}

const diffPatchStr = `diff --git a/README.md b/README.md
index c8bdd38..47a0095 100644
--- a/README.md
+++ b/README.md
@@ -2,4 +2,4 @@
 
 Blah Blah Blah this is a test 123
 
-Hello Tom Cruise 4242-4242-4242-4242
+Hello John Wick
diff --git a/blah.txt b/blah.txt
new file mode 100644
index 0000000..e9ea42a
--- /dev/null
+++ b/blah.txt
@@ -0,0 +1 @@
+this is a text file
diff --git a/main.go b/main.go
index e0fe924..0405bc6 100644
--- a/main.go
+++ b/main.go
@@ -3,5 +3,5 @@ package TestRepo
 import "fmt"
 
 func main() {
-	fmt.Println("This is a test")
+	fmt.Println("This is a test: My name is Tom Cruise")
+
 }`

var expectedFileDiff1 = &diffreviewer.FileDiff{
	PathOld: "README.md",
	PathNew: "README.md",
	Hunks: []*diffreviewer.Hunk{{
		StartLineOld:  2,
		LineLengthOld: 4,
		StartLineNew:  2,
		LineLengthNew: 4,
		Lines: []*diffreviewer.Line{{
			Type:     diffreviewer.LineAdded,
			Content:  "Hello John Wick",
			LnumDiff: 5,
			LnumOld:  0,
			LnumNew:  5,
		}},
	}},
	Extended: []string{"diff --git a/README.md b/README.md", "index c8bdd38..47a0095 100644"},
}
var expectedFileDiff2 = &diffreviewer.FileDiff{
	PathOld: "/dev/null",
	PathNew: "blah.txt",
	Hunks: []*diffreviewer.Hunk{{
		StartLineOld:  0,
		LineLengthOld: 0,
		StartLineNew:  1,
		LineLengthNew: 1,
		Lines: []*diffreviewer.Line{{
			Type:     diffreviewer.LineAdded,
			Content:  "this is a text file",
			LnumDiff: 1,
			LnumOld:  0,
			LnumNew:  1,
		}},
	}},
	Extended: []string{"diff --git a/blah.txt b/blah.txt", "new file mode 100644", "index 0000000..e9ea42a"},
}
var expectedFileDiff3 = &diffreviewer.FileDiff{
	PathOld: "main.go",
	PathNew: "main.go",
	Hunks: []*diffreviewer.Hunk{{
		StartLineOld:  3,
		LineLengthOld: 5,
		StartLineNew:  3,
		LineLengthNew: 5,
		Section:       "package TestRepo",
		Lines: []*diffreviewer.Line{{
			Type: diffreviewer.LineAdded,
			Content: "	fmt.Println(\"This is a test: My name is Tom Cruise\")",
			LnumDiff: 5,
			LnumOld:  0,
			LnumNew:  6,
		}},
	}},
	Extended: []string{"diff --git a/main.go b/main.go", "index e0fe924..0405bc6 100644"},
}

func (gd *gitdiffTestSuite) TestGetDiff() {
	ctrl := gomock.NewController(gd.T())
	defer ctrl.Finish()

	repoURL := "git://github.com/nightfall_testing/running_test.git"
	baseRev := "1234"
	headRev := "56789"
	repoFilePath := "/test/repo/path"
	diffOpts := &gitdiff.DiffOptions{
		FilterLineType: map[diffreviewer.LineType]bool{
			diffreviewer.LineAdded: true,
		},
	}
	repo := &libgit2.Repository{}
	diff := &libgit2.Diff{}
	diffPatchBytes := []byte(diffPatchStr)
	mockLibgit := libgit_mock.NewLibgit(ctrl)

	gitdiff := gitdiff.Client{
		Libgit:       mockLibgit,
		RepoFilePath: repoFilePath,
	}

	expectedFileDiffs := []*diffreviewer.FileDiff{expectedFileDiff1, expectedFileDiff2, expectedFileDiff3}

	mockLibgit.EXPECT().Clone(repoURL, repoFilePath).Return(repo, nil)
	mockLibgit.EXPECT().DiffRevToRev(repo, baseRev, headRev).Return(diff, nil)
	mockLibgit.EXPECT().ConvertDiffToPatch(diff).Return(diffPatchBytes, nil)

	fileDiffs, err := gitdiff.GetDiff(baseRev, headRev, repoURL, diffOpts)
	gd.NoError(err, "Unexpected error in GetDiff")
	gd.Equal(expectedFileDiffs, fileDiffs, "Incorrect file diffs received from GetDiff")
}

func TestGitDiffClient(t *testing.T) {
	suite.Run(t, new(gitdiffTestSuite))
}

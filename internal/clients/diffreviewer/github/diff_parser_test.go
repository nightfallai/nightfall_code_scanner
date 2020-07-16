package github_test

import (
	"bytes"
	"testing"

	"github.com/watchtowerai/nightfall_dlp/internal/clients/diffreviewer"

	"github.com/stretchr/testify/suite"
	"github.com/watchtowerai/nightfall_dlp/internal/clients/diffreviewer/github"
)

type diffParserTestSuite struct {
	suite.Suite
}

const rawDiff = `diff --git a/blah2.txt b/blah2.txt
new file mode 100644
index 0000000..3b18e51
--- /dev/null
+++ b/blah2.txt
@@ -0,0 +1 @@
+hello world
diff --git a/main.go b/main.go
index 0405bc6..292efcc 100644
--- a/main.go
+++ b/main.go
@@ -3,5 +3,5 @@ package TestRepo
 import "fmt"
 
 func main() {
-	fmt.Println("This is a test: My name is Tom Cruise")
+	fmt.Println("This is a test: My name is Keanu Reeves")
 }`

var fileDiff1 = diffreviewer.FileDiff{
	PathOld: "/dev/null",
	PathNew: "b/blah2.txt",
	Hunks: []*diffreviewer.Hunk{{
		StartLineOld:  0,
		LineLengthOld: 0,
		StartLineNew:  1,
		LineLengthNew: 1,
		Lines: []*diffreviewer.Line{{
			Type:     diffreviewer.LineAdded,
			Content:  "hello world",
			LnumDiff: 1,
			LnumOld:  0,
			LnumNew:  1,
		}},
	}},
	Extended: []string{"diff --git a/blah2.txt b/blah2.txt", "new file mode 100644", "index 0000000..3b18e51"},
}

var fileDiff2 = diffreviewer.FileDiff{
	PathOld: "a/main.go",
	PathNew: "b/main.go",
	Hunks: []*diffreviewer.Hunk{{
		StartLineOld:  3,
		LineLengthOld: 5,
		StartLineNew:  3,
		LineLengthNew: 5,
		Section:       "package TestRepo",
		Lines: []*diffreviewer.Line{{
			Type:     diffreviewer.LineUnchanged,
			Content:  `import "fmt"`,
			LnumDiff: 1,
			LnumOld:  3,
			LnumNew:  3,
		}, {
			Type:     diffreviewer.LineUnchanged,
			Content:  "",
			LnumDiff: 2,
			LnumOld:  4,
			LnumNew:  4,
		}, {
			Type:     diffreviewer.LineUnchanged,
			Content:  "func main() {",
			LnumDiff: 3,
			LnumOld:  5,
			LnumNew:  5,
		}, {
			Type: diffreviewer.LineDeleted,
			Content: `	fmt.Println("This is a test: My name is Tom Cruise")`,
			LnumDiff: 4,
			LnumOld:  6,
			LnumNew:  0,
		}, {
			Type: diffreviewer.LineAdded,
			Content: `	fmt.Println("This is a test: My name is Keanu Reeves")`,
			LnumDiff: 5,
			LnumOld:  0,
			LnumNew:  6,
		}, {
			Type:     diffreviewer.LineUnchanged,
			Content:  "}",
			LnumDiff: 6,
			LnumOld:  7,
			LnumNew:  7,
		}},
	}},
	Extended: []string{"diff --git a/main.go b/main.go", "index 0405bc6..292efcc 100644"},
}

var expectedParsedFileDiffs = []*diffreviewer.FileDiff{&fileDiff1, &fileDiff2}

func (d *diffParserTestSuite) TestParseMultiFile() {
	fileDiffs, err := github.ParseMultiFile(bytes.NewReader([]byte(rawDiff)))
	d.NoError(err, "unexpected error in parse multi-file test")
	d.Equal(expectedParsedFileDiffs, fileDiffs, "invalid fileDiff return value")
}

func TestDiffParser(t *testing.T) {
	suite.Run(t, new(diffParserTestSuite))
}

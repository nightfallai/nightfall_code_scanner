package diffutils_test

import (
	"bytes"
	"testing"

	"github.com/nightfallai/nightfall_cli/internal/clients/diffreviewer/diffutils"

	"github.com/nightfallai/nightfall_cli/internal/clients/diffreviewer"
	"github.com/stretchr/testify/suite"
)

type diffParserTestSuite struct {
	suite.Suite
}

const rawDiff = `diff --git a/a/prefix.txt b/a/old_prefix.txt
similarity index 100%
rename from a/prefix.txt
rename to a/old_prefix.txt
diff --git a/b/new_prefix.txt b/b/new_prefix.txt
new file mode 100644
index 0000000..e9956ba
--- /dev/null
+++ b/b/new_prefix.txt
@@ -0,0 +1 @@
+/b prefix
diff --git a/b/prefix.txt b/b/prefix.txt
deleted file mode 100644
index da5f6ce..0000000
--- a/b/prefix.txt
+++ /dev/null
@@ -1 +0,0 @@
-b prefix
diff --git a/main.go b/main.go
index ebcbd89..cb5c356 100644
--- a/main.go
+++ b/main.go
@@ -3,6 +3,6 @@ package TestRepo
 import "fmt"
 
 func main() {
-	fmt.Println("This is a test: My name is Tom Cruise")
+	fmt.Println("This is a test: My name is Tom Cruise and I am on a mission")
 	fmt.Println("Hope the API works!")
 }`

var fileDiff1 = diffreviewer.FileDiff{
	Hunks: []*diffreviewer.Hunk{},
	Extended: []string{
		"diff --git a/a/prefix.txt b/a/old_prefix.txt",
		"similarity index 100%",
		"rename from a/prefix.txt",
		"rename to a/old_prefix.txt",
	},
}

var fileDiff2 = diffreviewer.FileDiff{
	PathOld: "/dev/null",
	PathNew: "b/new_prefix.txt",
	Hunks: []*diffreviewer.Hunk{{
		StartLineOld:  0,
		LineLengthOld: 0,
		StartLineNew:  1,
		LineLengthNew: 1,
		Lines: []*diffreviewer.Line{{
			Type:     diffreviewer.LineAdded,
			Content:  "/b prefix",
			LnumDiff: 1,
			LnumOld:  0,
			LnumNew:  1,
		}},
	}},
	Extended: []string{"diff --git a/b/new_prefix.txt b/b/new_prefix.txt", "new file mode 100644", "index 0000000..e9956ba"},
}

var fileDiff3 = diffreviewer.FileDiff{
	PathOld: "b/prefix.txt",
	PathNew: "/dev/null",
	Hunks: []*diffreviewer.Hunk{{
		StartLineOld:  1,
		LineLengthOld: 1,
		StartLineNew:  0,
		LineLengthNew: 0,
		Lines: []*diffreviewer.Line{{
			Type:     diffreviewer.LineDeleted,
			Content:  "b prefix",
			LnumDiff: 1,
			LnumOld:  1,
			LnumNew:  0,
		}},
	}},
	Extended: []string{"diff --git a/b/prefix.txt b/b/prefix.txt", "deleted file mode 100644", "index da5f6ce..0000000"},
}

var fileDiff4 = diffreviewer.FileDiff{
	PathOld: "main.go",
	PathNew: "main.go",
	Hunks: []*diffreviewer.Hunk{{
		StartLineOld:  3,
		LineLengthOld: 6,
		StartLineNew:  3,
		LineLengthNew: 6,
		Section:       "package TestRepo",
		Lines: []*diffreviewer.Line{
			{
				Type:     diffreviewer.LineUnchanged,
				Content:  `import "fmt"`,
				LnumDiff: 1,
				LnumOld:  3,
				LnumNew:  3,
			},
			{
				Type:     diffreviewer.LineUnchanged,
				Content:  "",
				LnumDiff: 2,
				LnumOld:  4,
				LnumNew:  4,
			},
			{
				Type:     diffreviewer.LineUnchanged,
				Content:  "func main() {",
				LnumDiff: 3,
				LnumOld:  5,
				LnumNew:  5,
			},
			{
				Type: diffreviewer.LineDeleted,
				Content: `	fmt.Println("This is a test: My name is Tom Cruise")`,
				LnumDiff: 4,
				LnumOld:  6,
				LnumNew:  0,
			},
			{
				Type: diffreviewer.LineAdded,
				Content: `	fmt.Println("This is a test: My name is Tom Cruise and I am on a mission")`,
				LnumDiff: 5,
				LnumOld:  0,
				LnumNew:  6,
			},
			{
				Type: diffreviewer.LineUnchanged,
				Content: `	fmt.Println("Hope the API works!")`,
				LnumDiff: 6,
				LnumOld:  7,
				LnumNew:  7,
			},
			{
				Type:     diffreviewer.LineUnchanged,
				Content:  "}",
				LnumDiff: 7,
				LnumOld:  8,
				LnumNew:  8,
			},
		},
	}},
	Extended: []string{"diff --git a/main.go b/main.go", "index ebcbd89..cb5c356 100644"},
}
var expectedParsedFileDiffs = []*diffreviewer.FileDiff{&fileDiff1, &fileDiff2, &fileDiff3, &fileDiff4}

func (d *diffParserTestSuite) TestParseMultiFile() {
	fileDiffs, err := diffutils.ParseMultiFile(bytes.NewReader([]byte(rawDiff)))
	d.NoError(err, "unexpected error in parse multi-file test")
	d.Equal(expectedParsedFileDiffs, fileDiffs, "invalid fileDiff return value")
}

func TestDiffParser(t *testing.T) {
	suite.Run(t, new(diffParserTestSuite))
}

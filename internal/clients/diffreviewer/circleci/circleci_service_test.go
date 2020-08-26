package circleci

import (
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer"
	circlelogger "github.com/nightfallai/nightfall_code_scanner/internal/clients/logger/circle_logger"
	"github.com/nightfallai/nightfall_code_scanner/internal/mocks/clients/gitdiff_mock"
	"github.com/nightfallai/nightfall_code_scanner/internal/nightfallconfig"
	nightfallAPI "github.com/nightfallai/nightfall_go_client/generated"
	"github.com/stretchr/testify/suite"
)

const expectedDiffResponseStr = `diff --git a/README.md b/README.md
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
var expectedFileDiffs = []*diffreviewer.FileDiff{expectedFileDiff1, expectedFileDiff2, expectedFileDiff3}
var circleLogger = circlelogger.NewDefaultCircleLogger()

type circleCiTestSuite struct {
	suite.Suite
}

type testParams struct {
	ctrl *gomock.Controller
	cs   *Service
	w    *httptest.ResponseRecorder
}

func (c *circleCiTestSuite) initTestParams() *testParams {
	tp := &testParams{}
	tp.ctrl = gomock.NewController(c.T())
	tp.w = httptest.NewRecorder()
	tp.cs = &Service{
		Logger: circleLogger,
	}
	return tp
}

const commitSha = "7b46da6e4d3259b1a1c470ee468e2cb3d9733802"
const prevCommitSha = "15bf9548d16caff9f398b5aae78a611fc60d55bd"
const testConfigFileName = "nightfall_test_config.json"
const excludedCreditCardRegex = "4242-4242-4242-[0-9]{4}"
const excludedApiToken = "xG0Ct4Wsu3OTcJnE1dFLAQfRgL6b8tIv"
const excludedIPRegex = "^127\\."

var envVars = []string{
	WorkspacePathEnvVar,
	NightfallAPIKeyEnvVar,
	CircleCurrentCommitShaEnvVar,
	CircleBeforeCommitEnvVar,
}

func (c *circleCiTestSuite) AfterTest(suiteName, testName string) {
	for _, e := range envVars {
		err := os.Unsetenv(e)
		c.NoErrorf(err, "Error unsetting var %s", e)
	}
}

func (c *circleCiTestSuite) TestLoadConfig() {
	tp := c.initTestParams()
	apiKey := "api-key"
	cc := nightfallAPI.CREDIT_CARD_NUMBER
	phone := nightfallAPI.PHONE_NUMBER
	ip := nightfallAPI.IP_ADDRESS
	workspace, err := os.Getwd()
	c.NoError(err, "Error getting workspace")
	workspacePath := path.Join(workspace, "../../../../test/data")
	os.Setenv(WorkspacePathEnvVar, workspacePath)
	os.Setenv(CircleCurrentCommitShaEnvVar, commitSha)
	os.Setenv(CircleBeforeCommitEnvVar, prevCommitSha)
	os.Setenv(NightfallAPIKeyEnvVar, apiKey)

	expectedNightfallConfig := &nightfallconfig.Config{
		NightfallAPIKey:            apiKey,
		NightfallDetectors:         []*nightfallAPI.Detector{&cc, &phone, &ip},
		NightfallMaxNumberRoutines: 20,
		TokenExclusionList:         []string{excludedCreditCardRegex, excludedApiToken, excludedIPRegex},
		FileInclusionList:          []string{"*"},
		FileExclusionList:          []string{".nightfalldlp/config.json"},
	}

	nightfallConfig, err := tp.cs.LoadConfig(testConfigFileName)
	c.NoError(err, "Error in LoadConfig")
	c.Equal(expectedNightfallConfig, nightfallConfig, "Incorrect nightfall config")
}

func (c *circleCiTestSuite) TestLoadConfigMissingApiKey() {
	tp := c.initTestParams()
	workspace, err := os.Getwd()
	c.NoError(err, "Error getting workspace")
	workspacePath := path.Join(workspace, "../../../../test/data")
	os.Setenv(WorkspacePathEnvVar, workspacePath)
	os.Setenv(CircleCurrentCommitShaEnvVar, commitSha)
	os.Setenv(CircleBeforeCommitEnvVar, prevCommitSha)

	_, err = tp.cs.LoadConfig(testConfigFileName)
	c.EqualError(
		err,
		"missing env var for nightfall api key",
		"incorrect error from missing api key test",
	)
}

func (c *circleCiTestSuite) TestGetDiff() {
	tp := c.initTestParams()
	ctrl := gomock.NewController(c.T())
	defer ctrl.Finish()
	mockGitDiff := gitdiff_mock.NewGitDiff(ctrl)
	tp.cs.GitDiff = mockGitDiff

	mockGitDiff.EXPECT().GetDiff().Return(expectedDiffResponseStr, nil)

	fileDiffs, err := tp.cs.GetDiff()
	c.NoError(err, "unexpected error in GetDiff")
	c.Equal(expectedFileDiffs, fileDiffs, "invalid fileDiff return value")
}

func (c *circleCiTestSuite) TestWriteComments() {
	//TODO: implement
}

func TestCircleCiClient(t *testing.T) {
	suite.Run(t, new(circleCiTestSuite))
}

package github_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v31/github"
	"github.com/stretchr/testify/suite"
	nightfallAPI "github.com/watchtowerai/nightfall_api/generated"
	"github.com/watchtowerai/nightfall_dlp/internal/clients/diffreviewer"
	githubservice "github.com/watchtowerai/nightfall_dlp/internal/clients/diffreviewer/github"
	mock "github.com/watchtowerai/nightfall_dlp/internal/mocks/clients"
	"github.com/watchtowerai/nightfall_dlp/internal/mocks/clients/githubapi_mock"
	"github.com/watchtowerai/nightfall_dlp/internal/nightfallconfig"
)

const getDiffTestUrl = "/repos/alan20854/TestRepo/pulls/2.diff"
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
 }`

var expectedFileDiff1 = &diffreviewer.FileDiff{
	PathOld: "a/README.md",
	PathNew: "b/README.md",
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
	PathNew: "b/blah.txt",
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
	PathOld: "a/main.go",
	PathNew: "b/main.go",
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

var testPRCheckRequest = &githubservice.CheckRequest{
	Owner:       "alan20854",
	Repo:        "TestRepo",
	PullRequest: 2,
	SHA:         "7b46da6e4d3259b1a1c470ee468e2cb3d9733802",
}

type githubTestSuite struct {
	suite.Suite
}

type testParams struct {
	ctrl *gomock.Controller
	gc   *githubservice.Service
	w    *httptest.ResponseRecorder
}

func (g *githubTestSuite) initTestParams() *testParams {
	tp := &testParams{}
	tp.ctrl = gomock.NewController(g.T())
	tp.w = httptest.NewRecorder()
	return tp
}

const testConfigFileName = "nightfall_test_config.json"

var envVars = []string{
	githubservice.WorkspacePathEnvVar,
	githubservice.RepoEnvVar,
	githubservice.CommitShaEnvVar,
	githubservice.NightfallAPIKeyEnvVar,
}

func (g *githubTestSuite) AfterTest(suiteName, testName string) {
	for _, e := range envVars {
		err := os.Unsetenv(e)
		g.NoErrorf(err, "Error unsetting var %s", e)
	}
}

func (g *githubTestSuite) TestLoadConfig() {
	apiKey := "api-key"
	sha := "1234"
	owner := "nightfallai"
	repo := "testRepo"
	repoFullName := fmt.Sprintf("%s/%s", owner, repo)
	workspace, err := os.Getwd()
	g.NoError(err, "Error getting workspace")
	workspacePath := path.Join(workspace, "../../../../test/data")
	os.Setenv(githubservice.WorkspacePathEnvVar, workspacePath)
	os.Setenv(githubservice.RepoEnvVar, repoFullName)
	os.Setenv(githubservice.CommitShaEnvVar, sha)
	os.Setenv(githubservice.NightfallAPIKeyEnvVar, apiKey)

	expectedNightfallConfig := &nightfallconfig.Config{
		NightfallAPIKey: apiKey,
		NightfallDetectors: nightfallconfig.DetectorConfig{
			nightfallAPI.CREDIT_CARD_NUMBER: nightfallAPI.POSSIBLE,
			nightfallAPI.PHONE_NUMBER:       nightfallAPI.LIKELY,
		},
	}
	expectedGithubCheckRequest := &githubservice.CheckRequest{
		Owner: owner,
		Repo:  repo,
		SHA:   sha,
	}

	diffReviewer := githubservice.NewGithubService(&http.Client{})
	nightfallConfig, err := diffReviewer.LoadConfig(testConfigFileName)
	g.NoError(err, "Error in LoadConfig")
	gh, ok := diffReviewer.(*githubservice.Service)
	g.Equal(true, ok, "Error casting to github.Client")
	g.Equal(expectedNightfallConfig, nightfallConfig, "Incorrect nightfall config")
	g.Equal(expectedGithubCheckRequest, gh.CheckRequest, "Incorrect nightfall config")
}

func (g *githubTestSuite) TestGetDiff() {
	tp := g.initTestParams()
	ctrl := gomock.NewController(g.T())
	defer ctrl.Finish()
	mockAPI := githubapi_mock.NewGithubAPI(tp.ctrl)
	testGithubService := &githubservice.Service{
		Client:       mockAPI,
		CheckRequest: testPRCheckRequest,
	}
	tp.gc = testGithubService
	mockResponseStr := expectedDiffResponseStr

	mockHTTPClient := mock.NewHTTPClient(tp.ctrl)
	mockHTTPClient.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(request *http.Request) (*http.Response, error) {
			g.Equal(getDiffTestUrl, request.URL)
			resp := http.Response{
				Body: ioutil.NopCloser(strings.NewReader(expectedDiffResponseStr)),
			}
			return &resp, nil
		})
	opts := github.RawOptions{Type: github.Diff}
	mockAPI.EXPECT().
		GetRaw(
			context.Background(),
			testPRCheckRequest.Owner,
			testPRCheckRequest.Repo,
			testPRCheckRequest.PullRequest,
			opts,
		).Return(mockResponseStr, nil, nil)
	fileDiffs, err := tp.gc.GetDiff()
	g.NoError(err, "unexpected error in GetDiff")
	g.Equal(expectedFileDiffs, fileDiffs, "invalid fileDiff return value")
}

func TestGithubClient(t *testing.T) {
	suite.Run(t, new(githubTestSuite))
}

package github_test

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v31/github"
	"github.com/stretchr/testify/suite"
	nightfallAPI "github.com/watchtowerai/nightfall_api/generated"
	"github.com/watchtowerai/nightfall_dlp/internal/clients/diffreviewer"
	githubservice "github.com/watchtowerai/nightfall_dlp/internal/clients/diffreviewer/github"
	"github.com/watchtowerai/nightfall_dlp/internal/mocks/clients/githubapi_mock"
	"github.com/watchtowerai/nightfall_dlp/internal/nightfallconfig"
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

var testPRCheckRequest = &githubservice.CheckRequest{
	Owner:       "alan20854",
	Repo:        "TestRepo",
	PullRequest: 2,
	SHA:         "7b46da6e4d3259b1a1c470ee468e2cb3d9733802",
}
var testPRCheckRequestNoPR = &githubservice.CheckRequest{
	Owner: "alan20854",
	Repo:  "TestRepo",
	SHA:   "7b46da6e4d3259b1a1c470ee468e2cb3d9733802",
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
	githubservice.EventPathEnvVar,
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
	pullRequest := 1
	workspace, err := os.Getwd()
	g.NoError(err, "Error getting workspace")
	workspacePath := path.Join(workspace, "../../../../test/data")
	eventPath := path.Join(workspace, "../../../../test/data/github_action_event.json")
	os.Setenv(githubservice.WorkspacePathEnvVar, workspacePath)
	os.Setenv(githubservice.EventPathEnvVar, eventPath)
	os.Setenv(githubservice.NightfallAPIKeyEnvVar, apiKey)

	expectedNightfallConfig := &nightfallconfig.Config{
		NightfallAPIKey: apiKey,
		NightfallDetectors: nightfallconfig.DetectorConfig{
			nightfallAPI.CREDIT_CARD_NUMBER: nightfallAPI.POSSIBLE,
			nightfallAPI.PHONE_NUMBER:       nightfallAPI.LIKELY,
		},
	}
	expectedGithubCheckRequest := &githubservice.CheckRequest{
		Owner:       owner,
		Repo:        repo,
		SHA:         sha,
		PullRequest: pullRequest,
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

	tests := []struct {
		haveCheckRequest *githubservice.CheckRequest
		haveRawResponse  string
		wantFileDiffs    []*diffreviewer.FileDiff
	}{
		{
			haveCheckRequest: testPRCheckRequest,
			haveRawResponse:  expectedDiffResponseStr,
			wantFileDiffs:    expectedFileDiffs,
		},
		{
			haveCheckRequest: testPRCheckRequestNoPR,
			haveRawResponse:  expectedDiffResponseStr,
			wantFileDiffs:    expectedFileDiffs,
		},
	}

	for _, tt := range tests {
		testGithubService := &githubservice.Service{
			Client:       mockAPI,
			CheckRequest: tt.haveCheckRequest,
		}
		tp.gc = testGithubService
		baseBranch := "master"
		tp.gc.BaseBranch = baseBranch
		if tt.haveCheckRequest.PullRequest == 0 {
			mockAPI.EXPECT().
				GetRawBySha(
					context.Background(),
					testPRCheckRequest.Owner,
					testPRCheckRequest.Repo,
					testPRCheckRequest.SHA,
					baseBranch,
				).Return(tt.haveRawResponse, nil, nil)
		} else {
			opts := github.RawOptions{Type: github.Diff}
			mockAPI.EXPECT().
				GetRaw(
					context.Background(),
					testPRCheckRequest.Owner,
					testPRCheckRequest.Repo,
					testPRCheckRequest.PullRequest,
					opts,
				).Return(tt.haveRawResponse, nil, nil)
		}
		fileDiffs, err := tp.gc.GetDiff()
		g.NoError(err, "unexpected error in GetDiff")
		g.Equal(tt.wantFileDiffs, fileDiffs, "invalid fileDiff return value")
	}
}

func (g *githubTestSuite) TestWriteComments() {
	tp := g.initTestParams()
	ctrl := gomock.NewController(g.T())
	defer ctrl.Finish()
	mockAPI := githubapi_mock.NewGithubAPI(tp.ctrl)
	testGithubService := &githubservice.Service{
		Client:       mockAPI,
		CheckRequest: testPRCheckRequest,
	}
	tp.gc = testGithubService

	singleBatchComments, singleBatchAnnotations := makeTestCommentsAndAnnotations(
		"testComment",
		"/comments.txt",
		10,
	)
	multiBatchComments, multiBatchAnnotations := makeTestCommentsAndAnnotations(
		"testComment",
		"/comments.txt",
		70,
	)
	emptyComments, emptyAnnotations := []*diffreviewer.Comment{}, []*github.CheckRunAnnotation{}

	failureConclusion := "failure"
	successConclusion := "success"

	tests := []struct {
		giveComments    []*diffreviewer.Comment
		wantAnnotations []*github.CheckRunAnnotation
		wantConclusion  string
		desc            string
	}{
		{
			giveComments:    singleBatchComments,
			wantAnnotations: singleBatchAnnotations,
			wantConclusion:  failureConclusion,
			desc:            "single batch comments test",
		},
		{
			giveComments:    multiBatchComments,
			wantAnnotations: multiBatchAnnotations,
			wantConclusion:  failureConclusion,
			desc:            "multiple batch comments test",
		},
		{
			giveComments:    emptyComments,
			wantAnnotations: emptyAnnotations,
			wantConclusion:  successConclusion,
			desc:            "no comments test",
		},
	}

	checkName := "NightfallDLP"
	anotherCheckName := "build_test"
	checkRunInProgress := "in_progress"
	githubAction := "Github Actions"
	circleCI := "Circle CI"

	expectedCheckRunID := github.Int64(879322521)
	expectedCheckRun := github.CheckRun{
		ID:      expectedCheckRunID,
		HeadSHA: &testPRCheckRequest.SHA,
		Status:  &checkRunInProgress,
		Name:    &checkName,
		App: &github.App{
			Name: &githubAction,
		},
	}
	anotherCheckRun := github.CheckRun{
		ID:      expectedCheckRunID,
		HeadSHA: &testPRCheckRequest.SHA,
		Status:  &checkRunInProgress,
		Name:    &anotherCheckName,
		App: &github.App{
			Name: &circleCI,
		},
	}
	checkRuns := []*github.CheckRun{
		&anotherCheckRun,
		&expectedCheckRun,
	}
	totalCheckRuns := len(checkRuns)
	expectedListCheckRuns := github.ListCheckRunsResults{
		Total:     &totalCheckRuns,
		CheckRuns: checkRuns,
	}

	for _, tt := range tests {
		listCheckRunsOpt := github.ListCheckRunsOptions{Status: &checkRunInProgress}
		mockAPI.EXPECT().ListCheckRunsForRef(
			context.Background(),
			testPRCheckRequest.Owner,
			testPRCheckRequest.Repo,
			testPRCheckRequest.SHA,
			&listCheckRunsOpt,
		).Return(&expectedListCheckRuns, nil, nil)

		annotations := tt.wantAnnotations

		numUpdateRequests := int(math.Ceil(float64(len(annotations)) / githubservice.MaxAnnotationsPerRequest))
		for i := 0; i < numUpdateRequests; i++ {
			startCommentIdx := i * githubservice.MaxAnnotationsPerRequest
			endCommentIdx := min(startCommentIdx+githubservice.MaxAnnotationsPerRequest, len(annotations))
			updateOpt := github.UpdateCheckRunOptions{
				Name: checkName,
				Output: &github.CheckRunOutput{
					Title:       &checkName,
					Annotations: annotations[startCommentIdx:endCommentIdx],
					Summary:     github.String(""),
				},
			}
			expectedUpdatedCheckRun := &github.CheckRun{
				Output: updateOpt.Output,
				Name:   expectedCheckRun.Name,
			}
			mockAPI.EXPECT().UpdateCheckRun(
				context.Background(),
				testPRCheckRequest.Owner,
				testPRCheckRequest.Repo,
				*expectedCheckRun.ID,
				updateOpt,
			).Return(expectedUpdatedCheckRun, nil, nil)
		}

		checkRunCompletedStatus := "completed"
		completedOpt := github.UpdateCheckRunOptions{
			Status:     &checkRunCompletedStatus,
			Conclusion: &tt.wantConclusion,
		}
		mockAPI.EXPECT().UpdateCheckRun(
			context.Background(),
			testPRCheckRequest.Owner,
			testPRCheckRequest.Repo,
			*expectedCheckRunID,
			completedOpt,
		)

		err := tp.gc.WriteComments(tt.giveComments)
		g.NoError(err, fmt.Sprintf("Error getting actions dump for %s test", tt.desc))
	}
}

func makeTestCommentsAndAnnotations(body, filePath string, size int) ([]*diffreviewer.Comment, []*github.CheckRunAnnotation) {
	comments := make([]*diffreviewer.Comment, size)
	annotations := make([]*github.CheckRunAnnotation, size)
	annotationLevelFailure := "failure"
	for i := 0; i < size; i++ {
		comments[i] = &diffreviewer.Comment{
			Body:       body,
			FilePath:   filePath,
			LineNumber: i + 1,
		}
		annotations[i] = &github.CheckRunAnnotation{
			Path:            &filePath,
			StartLine:       &comments[i].LineNumber,
			EndLine:         &comments[i].LineNumber,
			AnnotationLevel: &annotationLevelFailure,
			Message:         &body,
		}
	}
	return comments, annotations
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func TestGithubClient(t *testing.T) {
	suite.Run(t, new(githubTestSuite))
}

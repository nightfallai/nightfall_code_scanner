package github_test

import (
	"context"
	"fmt"
	"math"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v31/github"
	"github.com/nightfallai/jenkins_test/internal/clients/diffreviewer"
	githubservice "github.com/nightfallai/jenkins_test/internal/clients/diffreviewer/github"
	githublogger "github.com/nightfallai/jenkins_test/internal/clients/logger/github_logger"
	"github.com/nightfallai/jenkins_test/internal/mocks/clients/githubchecks_mock"
	"github.com/nightfallai/jenkins_test/internal/mocks/clients/githubclient_mock"
	"github.com/nightfallai/jenkins_test/internal/nightfallconfig"
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
 }`

var logger = githublogger.NewDefaultGithubLogger()
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
	tp.gc = &githubservice.Service{
		Logger: logger,
	}
	return tp
}

const testConfigFileName = "nightfall_test_config.json"
const excludedCreditCard = "4242-4242-4242-4242"
const excludedApiToken = "xG0Ct4Wsu3OTcJnE1dFLAQfRgL6b8tIv"

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
	tp := g.initTestParams()
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
		TokenExclusionList: []string{excludedCreditCard, excludedApiToken},
	}
	expectedGithubCheckRequest := &githubservice.CheckRequest{
		Owner:       owner,
		Repo:        repo,
		SHA:         sha,
		PullRequest: pullRequest,
	}

	nightfallConfig, err := tp.gc.LoadConfig(testConfigFileName)
	g.NoError(err, "Error in LoadConfig")
	g.Equal(expectedNightfallConfig, nightfallConfig, "Incorrect nightfall config")
	g.Equal(expectedGithubCheckRequest, tp.gc.CheckRequest, "Incorrect nightfall config")
}

func (g *githubTestSuite) TestGetDiff() {
	tp := g.initTestParams()
	testDiffFilePath := githubservice.NightfallDiffFileName
	f, _ := os.Create(testDiffFilePath)
	defer os.Remove(f.Name())
	_, _ = f.Write([]byte(expectedDiffResponseStr))
	fileDiffs, err := tp.gc.GetDiff()
	g.NoError(err, "unexpected error in GetDiff")
	g.Equal(expectedFileDiffs, fileDiffs, "invalid fileDiff return value")
}

func (g *githubTestSuite) TestWriteComments() {
	tp := g.initTestParams()
	ctrl := gomock.NewController(g.T())
	defer ctrl.Finish()
	mockClient := githubclient_mock.NewGithubClient(tp.ctrl)
	mockChecks := githubchecks_mock.NewGithubChecks(tp.ctrl)
	testGithubService := &githubservice.Service{
		Client:       mockClient,
		CheckRequest: testPRCheckRequest,
		Logger:       logger,
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
		120,
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

	imageURL := "https://www.finsmes.com/wp-content/uploads/2019/11/Nightfall-AI.png"
	imageAlt := "Nightfall Logo"
	checkName := "Nightfall DLP"
	checkRunInProgressStatus := "in_progress"
	checkRunCompletedStatus := "completed"
	checkRunInProgress := "in_progress"
	createOpt := github.CreateCheckRunOptions{
		Name:    checkName,
		HeadSHA: testPRCheckRequest.SHA,
		Status:  &checkRunInProgress,
	}

	expectedCheckRunID := github.Int64(879322521)
	expectedCheckRun := github.CheckRun{
		ID:      expectedCheckRunID,
		HeadSHA: &testPRCheckRequest.SHA,
		Status:  &checkRunInProgressStatus,
		Name:    &checkName,
	}
	for _, tt := range tests {
		mockClient.EXPECT().ChecksService().Return(mockChecks)
		mockChecks.EXPECT().CreateCheckRun(
			context.Background(),
			testPRCheckRequest.Owner,
			testPRCheckRequest.Repo,
			createOpt,
		).Return(&expectedCheckRun, nil, nil)

		annotations := tt.wantAnnotations
		annotationLength := len(annotations)
		summaryString := fmt.Sprintf("Nightfall DLP has found %d potentially sensitive items", annotationLength)
		if len(annotations) == 0 {
			successfulOpt := github.UpdateCheckRunOptions{
				Name:       checkName,
				Status:     &checkRunCompletedStatus,
				Conclusion: &tt.wantConclusion,
				Output: &github.CheckRunOutput{
					Title:   &checkName,
					Summary: github.String(summaryString),
					Images: []*github.CheckRunImage{
						&github.CheckRunImage{
							Alt:      github.String(imageAlt),
							ImageURL: github.String(imageURL),
						},
					},
				},
			}
			mockClient.EXPECT().ChecksService().Return(mockChecks)
			mockChecks.EXPECT().UpdateCheckRun(
				context.Background(),
				testPRCheckRequest.Owner,
				testPRCheckRequest.Repo,
				*expectedCheckRunID,
				successfulOpt,
			)
		} else {
			numUpdateRequests := int(math.Ceil(float64(len(tt.wantAnnotations)) / githubservice.MaxAnnotationsPerRequest))
			for i := 0; i < numUpdateRequests-1; i++ {
				startCommentIdx := i * githubservice.MaxAnnotationsPerRequest
				endCommentIdx := min(startCommentIdx+githubservice.MaxAnnotationsPerRequest, len(tt.wantAnnotations))
				updateOpt := github.UpdateCheckRunOptions{
					Name: checkName,
					Output: &github.CheckRunOutput{
						Title:       &checkName,
						Summary:     github.String(summaryString),
						Annotations: tt.wantAnnotations[startCommentIdx:endCommentIdx],
					},
				}
				expectedUpdatedCheckRun := &github.CheckRun{
					Output: updateOpt.Output,
					Name:   expectedCheckRun.Name,
				}
				mockClient.EXPECT().ChecksService().Return(mockChecks)
				mockChecks.EXPECT().UpdateCheckRun(
					context.Background(),
					testPRCheckRequest.Owner,
					testPRCheckRequest.Repo,
					*expectedCheckRun.ID,
					updateOpt,
				).Return(expectedUpdatedCheckRun, nil, nil)
			}
			lastAnnotations := annotations[(numUpdateRequests-1)*githubservice.MaxAnnotationsPerRequest:]
			lastUpdateOpt := github.UpdateCheckRunOptions{
				Name:       checkName,
				Status:     &checkRunCompletedStatus,
				Conclusion: &failureConclusion,
				Output: &github.CheckRunOutput{
					Title:       &checkName,
					Summary:     github.String(summaryString),
					Annotations: lastAnnotations,
					Images: []*github.CheckRunImage{
						&github.CheckRunImage{
							Alt:      github.String(imageAlt),
							ImageURL: github.String(imageURL),
						},
					},
				},
			}
			expectedLastUpdatedCheckRun := &github.CheckRun{
				Name:       expectedCheckRun.Name,
				Status:     &checkRunCompletedStatus,
				Conclusion: &failureConclusion,
				Output:     lastUpdateOpt.Output,
			}
			mockClient.EXPECT().ChecksService().Return(mockChecks)
			mockChecks.EXPECT().UpdateCheckRun(
				context.Background(),
				testPRCheckRequest.Owner,
				testPRCheckRequest.Repo,
				*expectedCheckRun.ID,
				lastUpdateOpt,
			).Return(expectedLastUpdatedCheckRun, nil, nil)
		}
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
			Title:      "title",
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
			Title:           &comments[i].Title,
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

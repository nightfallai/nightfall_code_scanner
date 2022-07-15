package github

import (
	"context"
	"fmt"
	"math"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v33/github"
	"github.com/google/uuid"
	nf "github.com/nightfallai/nightfall-go-sdk"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer"
	githublogger "github.com/nightfallai/nightfall_code_scanner/internal/clients/logger/github_logger"
	"github.com/nightfallai/nightfall_code_scanner/internal/mocks/clients/gitdiff_mock"
	"github.com/nightfallai/nightfall_code_scanner/internal/mocks/clients/githubchecks_mock"
	"github.com/nightfallai/nightfall_code_scanner/internal/mocks/clients/githubclient_mock"
	"github.com/nightfallai/nightfall_code_scanner/internal/nightfallconfig"
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

var log = githublogger.NewDefaultGithubLogger()
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

var testPRCheckRequest = &CheckRequest{
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
	gc   *Service
	w    *httptest.ResponseRecorder
}

func (g *githubTestSuite) initTestParams() *testParams {
	tp := &testParams{}
	tp.ctrl = gomock.NewController(g.T())
	tp.w = httptest.NewRecorder()
	tp.gc = &Service{
		Logger: log,
	}
	return tp
}

const testConfigFileName = "nightfall_test_config.json"
const testEmptyConfigFileName = "nightfall_test_empty_config.json"
const testConfigDetectionRuleUUIDFileName = "nightfall_test_config_detection_rule_uuid.json"
const excludedCreditCardRegex = "4242-4242-4242-[0-9]{4}"
const excludedApiToken = "xG0Ct4Wsu3OTcJnE1dFLAQfRgL6b8tIv"
const excludedIPRegex = "^127\\."

var envVars = []string{
	WorkspacePathEnvVar,
	EventPathEnvVar,
	NightfallAPIKeyEnvVar,
}

var testDetectionRuleUUID = uuid.MustParse("9c1fd2c9-8ef5-40c4-b661-bd750ff0d684")

func (g *githubTestSuite) AfterTest(_, _ string) {
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
	_ = os.Setenv(WorkspacePathEnvVar, workspacePath)
	_ = os.Setenv(EventPathEnvVar, eventPath)
	_ = os.Setenv(NightfallAPIKeyEnvVar, apiKey)

	expectedNightfallConfig := &nightfallconfig.Config{
		NightfallAPIKey: apiKey,
		NightfallDetectionRules: []nf.DetectionRule{
			{
				Name: "my detection rule",
				Detectors: []nf.Detector{
					{
						MinNumFindings:    1,
						MinConfidence:     nf.ConfidencePossible,
						DetectorType:      nf.DetectorTypeNightfallDetector,
						DisplayName:       "cc",
						NightfallDetector: "CREDIT_CARD_NUMBER",
					},
					{
						MinNumFindings:    1,
						MinConfidence:     nf.ConfidencePossible,
						DetectorType:      nf.DetectorTypeNightfallDetector,
						DisplayName:       "phone",
						NightfallDetector: "PHONE_NUMBER",
					},
					{
						MinNumFindings:    1,
						MinConfidence:     nf.ConfidenceLikely,
						DetectorType:      nf.DetectorTypeNightfallDetector,
						DisplayName:       "ip",
						NightfallDetector: "IP_ADDRESS",
					},
				},
				LogicalOp: nf.LogicalOpAny,
			},
		},
		NightfallMaxNumberRoutines: 20,
		TokenExclusionList:         []string{excludedCreditCardRegex, excludedApiToken, excludedIPRegex},
		FileInclusionList:          []string{"*"},
		FileExclusionList:          []string{".nightfalldlp/config.json"},
		DefaultRedactionConfig: &nf.RedactionConfig{
			SubstitutionConfig: &nf.SubstitutionConfig{SubstitutionPhrase: "REDACTED"},
		},
		AnnotationLevel: "warning",
	}
	expectedGithubCheckRequest := &CheckRequest{
		Owner:       owner,
		Repo:        repo,
		SHA:         sha,
		PullRequest: pullRequest,
	}

	nightfallConfig, err := tp.gc.LoadConfig(testConfigFileName)
	g.NoError(err, "Unexpected error in LoadConfig")
	g.Equal(expectedNightfallConfig, nightfallConfig, "Incorrect nightfall config")
	g.Equal(expectedGithubCheckRequest, tp.gc.CheckRequest, "Incorrect github check request")
}

func (g *githubTestSuite) TestLoadConfigDetectionRuleUUID() {
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
	_ = os.Setenv(WorkspacePathEnvVar, workspacePath)
	_ = os.Setenv(EventPathEnvVar, eventPath)
	_ = os.Setenv(NightfallAPIKeyEnvVar, apiKey)

	expectedNightfallConfig := &nightfallconfig.Config{
		NightfallAPIKey:             apiKey,
		NightfallDetectionRuleUUIDs: []uuid.UUID{testDetectionRuleUUID},
		NightfallMaxNumberRoutines:  20,
		TokenExclusionList:          []string{excludedCreditCardRegex, excludedApiToken, excludedIPRegex},
		FileInclusionList:           []string{"*"},
		FileExclusionList:           []string{".nightfalldlp/config.json"},
		AnnotationLevel:             "failure",
	}
	expectedGithubCheckRequest := &CheckRequest{
		Owner:       owner,
		Repo:        repo,
		SHA:         sha,
		PullRequest: pullRequest,
	}

	nightfallConfig, err := tp.gc.LoadConfig(testConfigDetectionRuleUUIDFileName)
	g.NoError(err, "Unexpected error in LoadConfig")
	g.Equal(expectedNightfallConfig, nightfallConfig, "Incorrect nightfall config")
	g.Equal(expectedGithubCheckRequest, tp.gc.CheckRequest, "Incorrect github check request")
}

func (g *githubTestSuite) TestLoadEmptyConfig() {
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
	_ = os.Setenv(WorkspacePathEnvVar, workspacePath)
	_ = os.Setenv(EventPathEnvVar, eventPath)
	_ = os.Setenv(NightfallAPIKeyEnvVar, apiKey)

	expectedNightfallConfig := &nightfallconfig.Config{
		NightfallAPIKey: apiKey,
		NightfallDetectionRules: []nf.DetectionRule{
			{
				Name: "default detection rule",
				Detectors: []nf.Detector{
					{
						DetectorType:      nf.DetectorTypeNightfallDetector,
						NightfallDetector: "API_KEY",
						DisplayName:       "API_KEY",
						MinConfidence:     nf.ConfidencePossible,
						MinNumFindings:    1,
					},
					{
						DetectorType:      nf.DetectorTypeNightfallDetector,
						NightfallDetector: "CRYPTOGRAPHIC_KEY",
						DisplayName:       "CRYPTOGRAPHIC_KEY",
						MinConfidence:     nf.ConfidencePossible,
						MinNumFindings:    1,
					},
					{
						DetectorType:      nf.DetectorTypeNightfallDetector,
						NightfallDetector: "PASSWORD_IN_CODE",
						DisplayName:       "PASSWORD_IN_CODE",
						MinConfidence:     nf.ConfidencePossible,
						MinNumFindings:    1,
					},
				},
				LogicalOp: nf.LogicalOpAny,
			},
		},
		NightfallMaxNumberRoutines: nightfallconfig.DefaultMaxNumberRoutines,
		DefaultRedactionConfig: &nf.RedactionConfig{
			MaskConfig: &nf.MaskConfig{
				MaskingChar:             "*",
				NumCharsToLeaveUnmasked: 2,
			},
		},
		AnnotationLevel: "failure",
	}
	expectedGithubCheckRequest := &CheckRequest{
		Owner:       owner,
		Repo:        repo,
		SHA:         sha,
		PullRequest: pullRequest,
	}

	nightfallConfig, err := tp.gc.LoadConfig(testEmptyConfigFileName)
	g.NoError(err, "Unexpected error in LoadConfig")
	g.Equal(expectedNightfallConfig, nightfallConfig, "Incorrect nightfall config")
	g.Equal(expectedGithubCheckRequest, tp.gc.CheckRequest, "Incorrect github check request")
}

func (g *githubTestSuite) TestGetDiff() {
	tp := g.initTestParams()
	ctrl := gomock.NewController(g.T())
	defer ctrl.Finish()
	mockGitDiff := gitdiff_mock.NewGitDiff(ctrl)
	tp.gc.GitDiff = mockGitDiff

	mockGitDiff.EXPECT().GetDiff().Return(expectedDiffResponseStr, nil)

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
	testGithubService := &Service{
		Client:       mockClient,
		CheckRequest: testPRCheckRequest,
		Logger:       log,
	}
	tp.gc = testGithubService

	warnSingleBatchComments, warnSingleBatchAnnotations := makeTestCommentsAndAnnotations(
		"testComment",
		"/comments.txt",
		"warning",
		10,
	)
	failSingleBatchComments, failSingleBatchAnnotations := makeTestCommentsAndAnnotations(
		"testComment",
		"/comments.txt",
		"failure",
		10,
	)
	warnMultiBatchComments, warnMultiBatchAnnotations := makeTestCommentsAndAnnotations(
		"testComment",
		"/comments.txt",
		"warning",
		120,
	)
	failMultiBatchComments, failMultiBatchAnnotations := makeTestCommentsAndAnnotations(
		"testComment",
		"/comments.txt",
		"failure",
		120,
	)
	emptyComments, emptyAnnotations := make([]*diffreviewer.Comment, 0), make([]*github.CheckRunAnnotation, 0)

	failureConclusion := "failure"
	neutralConclusion := "neutral"
	successConclusion := "success"

	tests := []struct {
		giveComments    []*diffreviewer.Comment
		wantAnnotations []*github.CheckRunAnnotation
		wantConclusion  string
		desc            string
		annotationLevel string
	}{
		{
			giveComments:    warnSingleBatchComments,
			wantAnnotations: warnSingleBatchAnnotations,
			wantConclusion:  neutralConclusion,
			desc:            "warn single batch comments test",
			annotationLevel: "warning",
		},
		{
			giveComments:    warnMultiBatchComments,
			wantAnnotations: warnMultiBatchAnnotations,
			wantConclusion:  neutralConclusion,
			desc:            "warn multiple batch comments test",
			annotationLevel: "warning",
		},
		{
			giveComments:    failSingleBatchComments,
			wantAnnotations: failSingleBatchAnnotations,
			wantConclusion:  failureConclusion,
			desc:            "fail single batch comments test",
			annotationLevel: "failure",
		},
		{
			giveComments:    failMultiBatchComments,
			wantAnnotations: failMultiBatchAnnotations,
			wantConclusion:  failureConclusion,
			desc:            "fail multiple batch comments test",
			annotationLevel: "failure",
		},
		{
			giveComments:    emptyComments,
			wantAnnotations: emptyAnnotations,
			wantConclusion:  successConclusion,
			desc:            "no comments test",
			annotationLevel: "failure",
		},
	}

	imageURL := "https://cdn.nightfall.ai/nightfall-dark-logo-tm.png"
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
						{
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
			numUpdateRequests := int(math.Ceil(float64(len(tt.wantAnnotations)) / MaxAnnotationsPerRequest))
			for i := 0; i < numUpdateRequests-1; i++ {
				startCommentIdx := i * MaxAnnotationsPerRequest
				endCommentIdx := min(startCommentIdx+MaxAnnotationsPerRequest, len(tt.wantAnnotations))
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
			lastAnnotations := annotations[(numUpdateRequests-1)*MaxAnnotationsPerRequest:]
			lastUpdateOpt := github.UpdateCheckRunOptions{
				Name:       checkName,
				Status:     &checkRunCompletedStatus,
				Conclusion: &tt.wantConclusion,
				Output: &github.CheckRunOutput{
					Title:       &checkName,
					Summary:     github.String(summaryString),
					Annotations: lastAnnotations,
					Images: []*github.CheckRunImage{
						{
							Alt:      github.String(imageAlt),
							ImageURL: github.String(imageURL),
						},
					},
				},
			}
			expectedLastUpdatedCheckRun := &github.CheckRun{
				Name:       expectedCheckRun.Name,
				Status:     &checkRunCompletedStatus,
				Conclusion: &tt.wantConclusion,
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
		err := tp.gc.WriteComments(tt.giveComments, tt.annotationLevel)
		g.NoError(err, fmt.Sprintf("Error writing comments for %s test", tt.desc))
	}
}

func makeTestCommentsAndAnnotations(body, filePath, annotationLevel string, size int) ([]*diffreviewer.Comment, []*github.CheckRunAnnotation) {
	comments := make([]*diffreviewer.Comment, size)
	annotations := make([]*github.CheckRunAnnotation, size)
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
			AnnotationLevel: &annotationLevel,
			Message:         &body,
			Title:           &comments[i].Title,
		}
	}
	return comments, annotations
}

func TestGithubClient(t *testing.T) {
	suite.Run(t, new(githubTestSuite))
}

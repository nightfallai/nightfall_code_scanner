package circleci

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v33/github"
	"github.com/google/uuid"
	nf "github.com/nightfallai/nightfall-go-sdk"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer"
	circlelogger "github.com/nightfallai/nightfall_code_scanner/internal/clients/logger/circle_logger"
	"github.com/nightfallai/nightfall_code_scanner/internal/mocks/clients/gitdiff_mock"
	"github.com/nightfallai/nightfall_code_scanner/internal/mocks/clients/githubclient_mock"
	"github.com/nightfallai/nightfall_code_scanner/internal/mocks/clients/githubpullrequests_mock"
	"github.com/nightfallai/nightfall_code_scanner/internal/mocks/clients/githubrepositories_mock"
	loggermock "github.com/nightfallai/nightfall_code_scanner/internal/mocks/logger"
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
const testBranch = "testBranch"
const testOwner = "alan20854"
const testRepo = "TestRepo"
const testPrUrl = "https://github.com/alan20854/CircleCiTest/pull/3"
const testConfigFileName = "nightfall_test_config.json"
const testEmptyConfigFileName = "nightfall_empty_test_config.json"
const testConfigDetectionRuleUUIDFileName = "nightfall_test_config_detection_rule_uuid.json"
const excludedCreditCardRegex = "4242-4242-4242-[0-9]{4}"
const excludedApiToken = "xG0Ct4Wsu3OTcJnE1dFLAQfRgL6b8tIv"
const excludedIPRegex = "^127\\."

var envVars = []string{
	WorkspacePathEnvVar,
	CircleCurrentCommitShaEnvVar,
	CircleBeforeCommitEnvVar,
	CircleBranchEnvVar,
	CircleOwnerNameEnvVar,
	CircleRepoNameEnvVar,
	CirclePullRequestUrlEnvVar,
	NightfallAPIKeyEnvVar,
}
var testDetectionRuleUUID = uuid.MustParse("9c1fd2c9-8ef5-40c4-b661-bd750ff0d684")

func (c *circleCiTestSuite) AfterTest(_, _ string) {
	for _, e := range envVars {
		err := os.Unsetenv(e)
		c.NoErrorf(err, "Error unsetting var %s", e)
	}
}

func (c *circleCiTestSuite) TestLoadConfig() {
	tp := c.initTestParams()
	apiKey := "api-key"
	workspace, err := os.Getwd()
	c.NoError(err, "Error getting workspace")
	workspacePath := path.Join(workspace, "../../../../test/data")
	_ = os.Setenv(WorkspacePathEnvVar, workspacePath)
	_ = os.Setenv(CircleCurrentCommitShaEnvVar, commitSha)
	_ = os.Setenv(CircleBeforeCommitEnvVar, prevCommitSha)
	_ = os.Setenv(CircleBranchEnvVar, testBranch)
	_ = os.Setenv(CircleOwnerNameEnvVar, testOwner)
	_ = os.Setenv(CircleRepoNameEnvVar, testRepo)
	_ = os.Setenv(CirclePullRequestUrlEnvVar, testPrUrl)
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
	}

	nightfallConfig, err := tp.cs.LoadConfig(testConfigFileName)
	c.NoError(err, "Unexpected error in LoadConfig")
	c.Equal(expectedNightfallConfig, nightfallConfig, "Incorrect nightfall config")
}

func (c *circleCiTestSuite) TestLoadConfigDetectionRuleUUID() {
	tp := c.initTestParams()
	apiKey := "api-key"
	workspace, err := os.Getwd()
	c.NoError(err, "Error getting workspace")
	workspacePath := path.Join(workspace, "../../../../test/data")
	_ = os.Setenv(WorkspacePathEnvVar, workspacePath)
	_ = os.Setenv(CircleCurrentCommitShaEnvVar, commitSha)
	_ = os.Setenv(CircleBeforeCommitEnvVar, prevCommitSha)
	_ = os.Setenv(CircleBranchEnvVar, testBranch)
	_ = os.Setenv(CircleOwnerNameEnvVar, testOwner)
	_ = os.Setenv(CircleRepoNameEnvVar, testRepo)
	_ = os.Setenv(CirclePullRequestUrlEnvVar, testPrUrl)
	_ = os.Setenv(NightfallAPIKeyEnvVar, apiKey)

	expectedNightfallConfig := &nightfallconfig.Config{
		NightfallAPIKey:             apiKey,
		NightfallDetectionRuleUUIDs: []uuid.UUID{testDetectionRuleUUID},
		NightfallDetectionRules:     nil,
		NightfallMaxNumberRoutines:  20,
		TokenExclusionList:          []string{excludedCreditCardRegex, excludedApiToken, excludedIPRegex},
		FileInclusionList:           []string{"*"},
		FileExclusionList:           []string{".nightfalldlp/config.json"},
	}

	nightfallConfig, err := tp.cs.LoadConfig(testConfigDetectionRuleUUIDFileName)
	c.NoError(err, "Unexpected error in LoadConfig")
	c.Equal(expectedNightfallConfig, nightfallConfig, "Incorrect nightfall config")
}

func (c *circleCiTestSuite) TestLoadConfigMissingApiKey() {
	tp := c.initTestParams()
	workspace, err := os.Getwd()
	c.NoError(err, "Error getting workspace")
	workspacePath := path.Join(workspace, "../../../../test/data")
	_ = os.Setenv(WorkspacePathEnvVar, workspacePath)
	_ = os.Setenv(CircleCurrentCommitShaEnvVar, commitSha)
	_ = os.Setenv(CircleBeforeCommitEnvVar, prevCommitSha)
	_ = os.Setenv(CircleBranchEnvVar, testBranch)
	_ = os.Setenv(CircleOwnerNameEnvVar, testOwner)
	_ = os.Setenv(CircleRepoNameEnvVar, testRepo)

	_, err = tp.cs.LoadConfig(testConfigFileName)
	c.EqualError(
		err,
		"missing env var for nightfall api key",
		"incorrect error from missing api key test",
	)
}

func (c *circleCiTestSuite) TestLoadEmptyConfig() {
	tp := c.initTestParams()
	apiKey := "api-key"
	workspace, err := os.Getwd()
	c.NoError(err, "Error getting workspace")
	workspacePath := path.Join(workspace, "../../../../test/data")
	_ = os.Setenv(WorkspacePathEnvVar, workspacePath)
	_ = os.Setenv(CircleCurrentCommitShaEnvVar, commitSha)
	_ = os.Setenv(CircleBeforeCommitEnvVar, prevCommitSha)
	_ = os.Setenv(CircleBranchEnvVar, testBranch)
	_ = os.Setenv(CircleOwnerNameEnvVar, testOwner)
	_ = os.Setenv(CircleRepoNameEnvVar, testRepo)
	_ = os.Setenv(CirclePullRequestUrlEnvVar, testPrUrl)
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
	}

	nightfallConfig, err := tp.cs.LoadConfig(testEmptyConfigFileName)
	c.NoError(err, "Unexpected error in LoadConfig")
	c.Equal(expectedNightfallConfig, nightfallConfig, "Incorrect nightfall config")
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

func (c *circleCiTestSuite) TestWriteCircleComments() {
	tp := c.initTestParams()
	ctrl := gomock.NewController(c.T())
	defer ctrl.Finish()
	mockLogger := loggermock.NewLogger(ctrl)
	testCircleService := &Service{
		Logger: mockLogger,
		PrDetails: prDetails{
			CommitSha: commitSha,
			Owner:     testOwner,
			Repo:      testRepo,
		},
	}
	tp.cs = testCircleService

	testComments, _ := makeTestGithubRepositoryComments(
		"testComment",
		"/comments.txt",
		tp.cs.PrDetails.CommitSha,
		60,
	)
	emptyComments := make([]*diffreviewer.Comment, 0)

	tests := []struct {
		giveComments []*diffreviewer.Comment
		wantErr      error
		desc         string
	}{
		{
			giveComments: testComments,
			wantErr:      errSensitiveItemsFound,
			desc:         "single batch comments test",
		},
		{
			giveComments: emptyComments,
			wantErr:      nil,
			desc:         "no comments test",
		},
	}

	for _, tt := range tests {
		if len(tt.giveComments) == 0 {
			mockLogger.EXPECT().Info("no sensitive items found")
		}
		for _, comment := range tt.giveComments {
			mockLogger.EXPECT().Error(fmt.Sprintf(
				"%s at %s on line %d",
				comment.Body,
				comment.FilePath,
				comment.LineNumber,
			))
		}
		err := tp.cs.WriteComments(tt.giveComments)
		if len(tt.giveComments) == 0 {
			c.NoError(err, fmt.Sprintf("unexpected error writing comments for %s test", tt.desc))
		} else {
			c.EqualError(err, tt.wantErr.Error(), fmt.Sprintf("invalid error writing comments for %s test", tt.desc))
		}
	}
}

func (c *circleCiTestSuite) TestWritePullRequestComments() {
	tp := c.initTestParams()
	ctrl := gomock.NewController(c.T())
	defer ctrl.Finish()
	mockClient := githubclient_mock.NewGithubClient(tp.ctrl)
	mockPullRequests := githubpullrequests_mock.NewGithubPullRequests(tp.ctrl)
	mockLogger := loggermock.NewLogger(ctrl)
	testCircleService := &Service{
		GithubClient: mockClient,
		Logger:       mockLogger,
		PrDetails: prDetails{
			CommitSha: commitSha,
			Owner:     testOwner,
			Repo:      testRepo,
			PrNumber:  github.Int(3),
		},
	}
	tp.cs = testCircleService

	testComments, testGithubComments := makeTestGithubPullRequestComments(
		"testComment",
		"/comments.txt",
		tp.cs.PrDetails.CommitSha,
		60,
	)
	emptyComments, emptyGithubComments := make([]*diffreviewer.Comment, 0), make([]*github.PullRequestComment, 0)

	tests := []struct {
		giveComments       []*diffreviewer.Comment
		giveGithubComments []*github.PullRequestComment
		wantError          error
		desc               string
	}{
		{
			giveComments:       testComments,
			giveGithubComments: testGithubComments,
			wantError:          errSensitiveItemsFound,
			desc:               "single batch comments test",
		},
		{
			giveComments:       emptyComments,
			giveGithubComments: emptyGithubComments,
			wantError:          nil,
			desc:               "no comments test",
		},
	}

	for _, tt := range tests {
		mockClient.EXPECT().PullRequestsService().Return(mockPullRequests)
		mockPullRequests.EXPECT().ListComments(
			context.Background(),
			testCircleService.PrDetails.Owner,
			testCircleService.PrDetails.Repo,
			*testCircleService.PrDetails.PrNumber,
			&github.PullRequestListCommentsOptions{},
		)
		if len(tt.giveComments) == 0 {
			mockLogger.EXPECT().Info("no sensitive items found")
		}
		for _, gc := range tt.giveGithubComments {
			mockClient.EXPECT().PullRequestsService().Return(mockPullRequests)
			mockPullRequests.EXPECT().CreateComment(
				context.Background(),
				testCircleService.PrDetails.Owner,
				testCircleService.PrDetails.Repo,
				*testCircleService.PrDetails.PrNumber,
				gc,
			)
			mockLogger.EXPECT().Error(fmt.Sprintf(
				"%s at %s on line %d",
				gc.GetBody(),
				gc.GetPath(),
				gc.GetLine(),
			))
		}
		err := tp.cs.WriteComments(tt.giveComments)
		if len(tt.giveComments) > 0 {
			c.EqualError(
				err,
				tt.wantError.Error(),
				fmt.Sprintf("invalid error writing comments for %s test", tt.desc),
			)
		} else {
			c.NoError(err, fmt.Sprintf("Error writing comments for %s test", tt.desc))
		}
	}
}

func makeTestGithubPullRequestComments(
	body,
	filePath,
	commitSha string,
	size int,
) ([]*diffreviewer.Comment, []*github.PullRequestComment) {
	comments := make([]*diffreviewer.Comment, size)
	githubComments := make([]*github.PullRequestComment, size)
	for i := 0; i < size; i++ {
		comments[i] = &diffreviewer.Comment{
			Title:      "title",
			Body:       body,
			FilePath:   filePath,
			LineNumber: i + 1,
		}
		githubComments[i] = &github.PullRequestComment{
			CommitID: &commitSha,
			Body:     &body,
			Path:     &filePath,
			Line:     &comments[i].LineNumber,
			Side:     github.String(GithubCommentRightSide),
		}
	}
	return comments, githubComments
}

func (c *circleCiTestSuite) TestWriteRepositoryComments() {
	tp := c.initTestParams()
	ctrl := gomock.NewController(c.T())
	defer ctrl.Finish()
	mockClient := githubclient_mock.NewGithubClient(tp.ctrl)
	mockRepositories := githubrepositories_mock.NewGithubRepositories(tp.ctrl)
	mockLogger := loggermock.NewLogger(ctrl)
	testCircleService := &Service{
		GithubClient: mockClient,
		Logger:       mockLogger,
		PrDetails: prDetails{
			CommitSha: commitSha,
			Owner:     testOwner,
			Repo:      testRepo,
		},
	}
	tp.cs = testCircleService

	testComments, testGithubComments := makeTestGithubRepositoryComments(
		"testComment",
		"/comments.txt",
		tp.cs.PrDetails.CommitSha,
		60,
	)
	emptyComments, emptyGithubComments := make([]*diffreviewer.Comment, 0), make([]*github.RepositoryComment, 0)

	tests := []struct {
		giveComments       []*diffreviewer.Comment
		giveGithubComments []*github.RepositoryComment
		wantError          error
		desc               string
	}{
		{
			giveComments:       testComments,
			giveGithubComments: testGithubComments,
			wantError:          errSensitiveItemsFound,
			desc:               "single batch comments test",
		},
		{
			giveComments:       emptyComments,
			giveGithubComments: emptyGithubComments,
			wantError:          nil,
			desc:               "no comments test",
		},
	}

	for _, tt := range tests {
		if len(tt.giveGithubComments) == 0 {
			mockLogger.EXPECT().Info("no sensitive items found")
		}
		for index, gc := range tt.giveGithubComments {
			mockClient.EXPECT().RepositoriesService().Return(mockRepositories)
			mockRepositories.EXPECT().CreateComment(
				context.Background(),
				testCircleService.PrDetails.Owner,
				testCircleService.PrDetails.Repo,
				testCircleService.PrDetails.CommitSha,
				gc,
			)
			mockLogger.EXPECT().Error(fmt.Sprintf(
				"%s at %s on line %d",
				gc.GetBody(),
				gc.GetPath(),
				tt.giveComments[index].LineNumber,
			))
		}
		err := tp.cs.WriteComments(tt.giveComments)
		if len(tt.giveComments) > 0 {
			c.EqualError(
				err,
				tt.wantError.Error(),
				fmt.Sprintf("invalid error writing comments for %s test", tt.desc),
			)
		} else {
			c.NoError(err, fmt.Sprintf("Error writing comments for %s test", tt.desc))
		}
	}
}

func makeTestGithubRepositoryComments(
	body,
	filePath,
	commitSha string,
	size int,
) ([]*diffreviewer.Comment, []*github.RepositoryComment) {
	comments := make([]*diffreviewer.Comment, size)
	githubComments := make([]*github.RepositoryComment, size)
	for i := 0; i < size; i++ {
		comments[i] = &diffreviewer.Comment{
			Title:      "title",
			Body:       body,
			FilePath:   filePath,
			LineNumber: i + 1,
		}
		githubComments[i] = &github.RepositoryComment{
			CommitID: &commitSha,
			Body:     &body,
			Path:     &filePath,
			Position: github.Int(i + 1),
		}
	}
	return comments, githubComments
}

func (c *circleCiTestSuite) TestFilterExistingComments() {
	bodyStrs := []string{"a", "b", "c"}
	pathStrs := []string{"a.txt", "b.txt", "c.txt"}
	lineNums := []int{1, 2, 3, 6}
	existingComments := []*github.PullRequestComment{
		{
			Body: &bodyStrs[0],
			Path: &pathStrs[0],
			Line: &lineNums[0],
		},
		{
			Body: &bodyStrs[1],
			Path: &pathStrs[1],
			Line: &lineNums[1],
		},
		{
			Body: &bodyStrs[2],
			Path: &pathStrs[2],
			Line: &lineNums[2],
		},
		{
			Body: &bodyStrs[0],
			Path: &pathStrs[0],
			Line: nil,
		},
	}
	newComment1 := &github.PullRequestComment{
		Body: &bodyStrs[0],
		Path: &pathStrs[0],
		Line: &lineNums[3],
	}
	newComment2 := &github.PullRequestComment{
		Body: &bodyStrs[1],
		Path: &pathStrs[1],
		Line: &lineNums[2],
	}
	comments := []*github.PullRequestComment{
		existingComments[0],
		existingComments[1],
		existingComments[2],
		newComment1,
		newComment2,
	}
	expectedFilteredComments := []*github.PullRequestComment{
		newComment1,
		newComment2,
	}
	filteredComments := filterExistingComments(comments, existingComments)
	c.Equal(expectedFilteredComments, filteredComments, "invalid filtered comments value")
}

func TestCircleCiClient(t *testing.T) {
	suite.Run(t, new(circleCiTestSuite))
}

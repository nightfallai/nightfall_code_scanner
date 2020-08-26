package circleci

import (
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	circlelogger "github.com/nightfallai/nightfall_code_scanner/internal/clients/logger/circle_logger"
	"github.com/nightfallai/nightfall_code_scanner/internal/nightfallconfig"
	nightfallAPI "github.com/nightfallai/nightfall_go_client/generated"
	"github.com/stretchr/testify/suite"
)

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
	//TODO: implement
}

func (c *circleCiTestSuite) TestWriteComments() {
	//TODO: implement
}

func TestCircleCiClient(t *testing.T) {
	suite.Run(t, new(circleCiTestSuite))
}

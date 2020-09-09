package nightfallconfig_test

import (
	"os"
	"path"
	"testing"

	githublogger "github.com/nightfallai/nightfall_code_scanner/internal/clients/logger/github_logger"
	"github.com/nightfallai/nightfall_code_scanner/internal/nightfallconfig"
	nightfallAPI "github.com/nightfallai/nightfall_go_client/generated"
	"github.com/stretchr/testify/assert"
)

const testFileName = "nightfall_test_config.json"
const testMissingFileName = "nightfall_test_missing_config.json"
const excludedCreditCardRegex = "4242-4242-4242-[0-9]{4}"
const excludedApiToken = "xG0Ct4Wsu3OTcJnE1dFLAQfRgL6b8tIv"
const excludedIPRegex = "^127\\."

func TestGetNightfallConfig(t *testing.T) {
	cc := nightfallAPI.CREDIT_CARD_NUMBER
	ip := nightfallAPI.IP_ADDRESS
	phone := nightfallAPI.PHONE_NUMBER
	workspaceConfig, err := os.Getwd()
	assert.NoError(t, err, "Unexpected error when getting current directory")
	workspacePath := path.Join(workspaceConfig, "../../test/data")
	expectedConfig := &nightfallconfig.NightfallConfigFileStructure{
		Detectors:          []*nightfallAPI.Detector{&cc, &phone, &ip},
		MaxNumberRoutines:  20,
		TokenExclusionList: []string{excludedCreditCardRegex, excludedApiToken, excludedIPRegex},
		FileInclusionList:  []string{"*"},
		FileExclusionList:  []string{".nightfalldlp/config.json"},
	}
	actualConfig := nightfallconfig.GetNightfallConfigFile(workspacePath, testFileName, nil)
	assert.Equal(t, expectedConfig, actualConfig, "Incorrect nightfall config")
}

func TestGetNightfallConfigMissingConfigFile(t *testing.T) {
	apiDetector := nightfallAPI.API_KEY
	cryptoDetector := nightfallAPI.CRYPTOGRAPHIC_TOKEN
	workspaceConfig, err := os.Getwd()
	assert.NoError(t, err, "Unexpected error when getting current directory")
	workspacePath := path.Join(workspaceConfig, "../../test/data")
	expectedConfig := &nightfallconfig.NightfallConfigFileStructure{
		Detectors:         []*nightfallAPI.Detector{&apiDetector, &cryptoDetector},
		MaxNumberRoutines: 20,
	}
	actualConfig := nightfallconfig.GetNightfallConfigFile(workspacePath, testMissingFileName, githublogger.NewDefaultGithubLogger())
	assert.Equal(t, expectedConfig, actualConfig, "Incorrect nightfall config")
}

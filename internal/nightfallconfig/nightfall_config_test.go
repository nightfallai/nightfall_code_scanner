package nightfallconfig_test

import (
	"encoding/json"
	"fmt"
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

var (
	one                           int32 = 1
	confidencePossible                  = nightfallAPI.CONFIDENCE_POSSIBLE
	ccDetector                          = nightfallAPI.NIGHTFALLDETECTORTYPE_CREDIT_CARD_NUMBER
	pnDetector                          = nightfallAPI.NIGHTFALLDETECTORTYPE_PHONE_NUMBER
	ipDetector                          = nightfallAPI.NIGHTFALLDETECTORTYPE_IP_ADDRESS
	nightfallDetectorType               = nightfallAPI.DETECTORTYPE_NIGHTFALL_DETECTOR
	nightfallAPIKey                     = nightfallAPI.NIGHTFALLDETECTORTYPE_API_KEY
	nightfallAPIKeyName                 = string(nightfallAPI.NIGHTFALLDETECTORTYPE_API_KEY)
	nightfallCryptographicKey           = nightfallAPI.NIGHTFALLDETECTORTYPE_CRYPTOGRAPHIC_KEY
	nightfallCryptographicKeyName       = string(nightfallAPI.NIGHTFALLDETECTORTYPE_CRYPTOGRAPHIC_KEY)
)

func TestGetNightfallConfig(t *testing.T) {
	workspaceConfig, err := os.Getwd()
	assert.NoError(t, err, "Unexpected error when getting current directory")
	workspacePath := path.Join(workspaceConfig, "../../test/data")
	expectedConfig := &nightfallconfig.NightfallConfigFileStructure{
		Conditions: []*nightfallAPI.Condition{
			{
				Detector: &nightfallAPI.Detector{
					DetectorType:      &nightfallDetectorType,
					NightfallDetector: &ccDetector,
				},
			},
			{
				Detector: &nightfallAPI.Detector{
					DetectorType:      &nightfallDetectorType,
					NightfallDetector: &pnDetector,
				},
			},
			{
				Detector: &nightfallAPI.Detector{
					DetectorType:      &nightfallDetectorType,
					NightfallDetector: &ipDetector,
				},
			},
		},
		MaxNumberRoutines:  20,
		TokenExclusionList: []string{excludedCreditCardRegex, excludedApiToken, excludedIPRegex},
		FileInclusionList:  []string{"*"},
		FileExclusionList:  []string{".nightfalldlp/config.json"},
	}
	actualConfig, err := nightfallconfig.GetNightfallConfigFile(workspacePath, testFileName, nil)
	val, _ := json.Marshal(expectedConfig)
	fmt.Println(string(val))
	assert.NoError(t, err, "Unexpected error in test GetNightfallConfig")
	assert.Equal(t, expectedConfig, actualConfig, "Incorrect nightfall config")
}

func TestGetNightfallConfigMissingConfigFile(t *testing.T) {
	workspaceConfig, err := os.Getwd()
	assert.NoError(t, err, "Unexpected error when getting current directory")
	workspacePath := path.Join(workspaceConfig, "../../test/data")
	expectedConfig := &nightfallconfig.NightfallConfigFileStructure{
		Conditions: []*nightfallAPI.Condition{
			{
				Detector: &nightfallAPI.Detector{
					DetectorType:      &nightfallDetectorType,
					NightfallDetector: &nightfallAPIKey,
					DisplayName:       &nightfallAPIKeyName,
				},
				MinNumFindings: &one,
				MinConfidence:  &confidencePossible,
			},
			{
				Detector: &nightfallAPI.Detector{
					DetectorType:      &nightfallDetectorType,
					NightfallDetector: &nightfallCryptographicKey,
					DisplayName:       &nightfallCryptographicKeyName,
				},
				MinNumFindings: &one,
				MinConfidence:  &confidencePossible,
			},
		},
		MaxNumberRoutines: nightfallconfig.DefaultMaxNumberRoutines,
	}
	actualConfig, err := nightfallconfig.GetNightfallConfigFile(workspacePath, testMissingFileName, githublogger.NewDefaultGithubLogger())
	assert.NoError(t, err, "Unexpected error in test GetNightfallConfigMissingConfigFile")
	assert.Equal(t, expectedConfig, actualConfig, "Incorrect nightfall config")
}

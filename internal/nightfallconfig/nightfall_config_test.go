package nightfallconfig

import (
	"os"
	"path"
	"testing"

	nf "github.com/nightfallai/nightfall-go-sdk"
	githublogger "github.com/nightfallai/nightfall_code_scanner/internal/clients/logger/github_logger"
	"github.com/stretchr/testify/assert"
)

const testFileName = "nightfall_test_config.json"
const testMissingFileName = "nightfall_test_missing_config.json"
const excludedCreditCardRegex = "4242-4242-4242-[0-9]{4}"
const excludedApiToken = "xG0Ct4Wsu3OTcJnE1dFLAQfRgL6b8tIv"
const excludedIPRegex = "^127\\."

func TestGetNightfallConfig(t *testing.T) {
	workspaceConfig, err := os.Getwd()
	assert.NoError(t, err, "Unexpected error when getting current directory")
	workspacePath := path.Join(workspaceConfig, "../../test/data")
	expectedConfig := &ConfigFile{
		DetectionRules: []nf.DetectionRule{
			{
				Name: "my detection rule",
				Detectors: []nf.Detector{
					{
						MinNumFindings:    1,
						MinConfidence:     nf.ConfidencePossible,
						DisplayName:       "cc",
						DetectorType:      nf.DetectorTypeNightfallDetector,
						NightfallDetector: "CREDIT_CARD_NUMBER",
					},
					{
						MinNumFindings:    1,
						MinConfidence:     nf.ConfidencePossible,
						DisplayName:       "phone",
						DetectorType:      nf.DetectorTypeNightfallDetector,
						NightfallDetector: "PHONE_NUMBER",
					},
					{
						MinNumFindings:    1,
						MinConfidence:     nf.ConfidenceLikely,
						DisplayName:       "ip",
						DetectorType:      nf.DetectorTypeNightfallDetector,
						NightfallDetector: "IP_ADDRESS",
					},
				},
				LogicalOp: nf.LogicalOpAny,
			},
		},
		MaxNumberRoutines:  20,
		TokenExclusionList: []string{excludedCreditCardRegex, excludedApiToken, excludedIPRegex},
		FileInclusionList:  []string{"*"},
		FileExclusionList:  []string{".nightfalldlp/config.json"},
		DefaultRedactionConfig: &nf.RedactionConfig{
			SubstitutionConfig: &nf.SubstitutionConfig{SubstitutionPhrase: "REDACTED"},
		},
	}
	actualConfig, err := GetNightfallConfigFile(workspacePath, testFileName, nil)
	assert.NoError(t, err, "Unexpected error in test GetNightfallConfig")
	assert.Equal(t, expectedConfig, actualConfig, "Incorrect nightfall config")
}

func TestGetNightfallConfigMissingConfigFile(t *testing.T) {
	workspaceConfig, err := os.Getwd()
	assert.NoError(t, err, "Unexpected error when getting current directory")
	workspacePath := path.Join(workspaceConfig, "../../test/data")
	expectedConfig := &ConfigFile{
		DetectionRules: []nf.DetectionRule{
			{
				Detectors: []nf.Detector{
					{
						MinNumFindings:    1,
						MinConfidence:     nf.ConfidencePossible,
						DisplayName:       "API_KEY",
						DetectorType:      nf.DetectorTypeNightfallDetector,
						NightfallDetector: "API_KEY",
					},
					{
						MinNumFindings:    1,
						MinConfidence:     nf.ConfidencePossible,
						DisplayName:       "CRYPTOGRAPHIC_KEY",
						DetectorType:      nf.DetectorTypeNightfallDetector,
						NightfallDetector: "CRYPTOGRAPHIC_KEY",
					},
					{
						MinNumFindings:    1,
						MinConfidence:     nf.ConfidencePossible,
						DisplayName:       "PASSWORD_IN_CODE",
						DetectorType:      nf.DetectorTypeNightfallDetector,
						NightfallDetector: "PASSWORD_IN_CODE",
					},
				},
			},
		},
		MaxNumberRoutines: DefaultMaxNumberRoutines,
	}
	actualConfig, err := GetNightfallConfigFile(workspacePath, testMissingFileName, githublogger.NewDefaultGithubLogger())
	assert.NoError(t, err, "Unexpected error in test GetNightfallConfigMissingConfigFile")
	assert.Equal(t, expectedConfig, actualConfig, "Incorrect nightfall config")
}

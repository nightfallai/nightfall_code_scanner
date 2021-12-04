package nightfallconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/google/uuid"
	nf "github.com/nightfallai/nightfall-go-sdk"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/logger"
)

const (
	// MaxConcurrentRoutinesCap is the maximum number of goroutines issuing concurrent requests to the Nightfall API
	MaxConcurrentRoutinesCap = 50
	DefaultMaxNumberRoutines = 20

	nightfallConfigFilename      = ".nightfalldlp/config.json"
	defaultConditionsInfoMessage = "Using default Detection Rule with detectors API_KEY and CRYPTOGRAPHIC_KEY"
)

var defaultNightfallConfig = &ConfigFile{
	DetectionRules: []nf.DetectionRule{
		{
			Detectors: []nf.Detector{
				{
					MinNumFindings:    1,
					MinConfidence:     nf.ConfidencePossible,
					DetectorType:      nf.DetectorTypeNightfallDetector,
					NightfallDetector: "API_KEY",
					DisplayName:       "API_KEY",
				},
				{
					MinNumFindings:    1,
					MinConfidence:     nf.ConfidencePossible,
					DetectorType:      nf.DetectorTypeNightfallDetector,
					NightfallDetector: "CRYPTOGRAPHIC_KEY",
					DisplayName:       "CRYPTOGRAPHIC_KEY",
				},
			},
		},
	},
	MaxNumberRoutines: DefaultMaxNumberRoutines,
}

// ConfigFile is the struct of the JSON nightfall config file
type ConfigFile struct {
	DetectionRuleUUIDs []uuid.UUID        `json:"detectionRuleUUIDs"`
	DetectionRules     []nf.DetectionRule `json:"detectionRules"`
	MaxNumberRoutines  int                `json:"maxNumberConcurrentRoutines"`
	TokenExclusionList []string           `json:"tokenExclusionList"`
	FileInclusionList  []string           `json:"fileInclusionList"`
	FileExclusionList  []string           `json:"fileExclusionList"`
}

// Config general config struct
type Config struct {
	NightfallAPIKey             string
	NightfallDetectionRuleUUIDs []uuid.UUID
	NightfallDetectionRules     []nf.DetectionRule
	NightfallMaxNumberRoutines  int
	TokenExclusionList          []string
	FileInclusionList           []string
	FileExclusionList           []string
}

// GetNightfallConfigFile loads nightfall config from file, returns default if missing/invalid
func GetNightfallConfigFile(workspacePath, fileName string, logger logger.Logger) (*ConfigFile, error) {
	nightfallConfigFile, err := os.Open(path.Join(workspacePath, fileName))
	if err != nil {
		logger.Warning(fmt.Sprintf("Error opening nightfall config: %s", err.Error()))
		logger.Info(defaultConditionsInfoMessage)
		return defaultNightfallConfig, nil
	}
	defer nightfallConfigFile.Close()
	byteValue, err := ioutil.ReadAll(nightfallConfigFile)
	if err != nil {
		logger.Warning(fmt.Sprintf("Error reading nightfall config: %s", err.Error()))
		logger.Info(defaultConditionsInfoMessage)
		return defaultNightfallConfig, nil
	}
	var nightfallConfig ConfigFile
	err = json.Unmarshal(byteValue, &nightfallConfig)
	if err != nil {
		return nil, err
	}
	if len(nightfallConfig.DetectionRuleUUIDs) < 1 && len(nightfallConfig.DetectionRules) < 1 {
		return nil, errors.New("nightfall config file is missing DetectionRuleUUIDs or inline DetectionRules")
	}
	if nightfallConfig.MaxNumberRoutines <= 0 {
		nightfallConfig.MaxNumberRoutines = DefaultMaxNumberRoutines
	} else if nightfallConfig.MaxNumberRoutines > MaxConcurrentRoutinesCap {
		nightfallConfig.MaxNumberRoutines = MaxConcurrentRoutinesCap
	}
	nightfallConfig.FileExclusionList = append(nightfallConfig.FileExclusionList, nightfallConfigFilename)
	return &nightfallConfig, nil
}

package nightfallconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/nightfallai/nightfall_code_scanner/internal/clients/logger"
	nightfallAPI "github.com/nightfallai/nightfall_go_client/generated"
)

// maximum number of routines (scan request + response) running at once
const MaxConcurrentRoutinesCap = 50
const DefaultMaxNumberRoutines = 20
const nightfallConfigFilename = ".nightfalldlp/config.json"
const defaultConditionsInfoMessage = "Using default Conditions with detectors API_KEY and CRYPTOGRAPHIC_KEY"

var defaultNightfallConfig = &NightfallConfigFileStructure{
	//Conditions: []*nightfallAPI.Condition{
	//	{
	//		Detector: nightfallAPI.Detector{
	//			DisplayName:       "",
	//			DetectorType:      "",
	//			NightfallDetector: nightfallAPI.NIGHTFALLDETECTORTYPE_API_KEY,
	//			Regex:             nightfallAPI.Regex{},
	//			WordList:          nightfallAPI.WordList{},
	//			ContextRules:      nil,
	//			ExclusionRules:    nil,
	//		},
	//	},
	//	{
	//		Detector: nightfallAPI.Detector{
	//			DisplayName:       "",
	//			DetectorType:      "",
	//			NightfallDetector: nightfallAPI.NIGHTFALLDETECTORTYPE_CRYPTOGRAPHIC_KEY,
	//			ContextRules:      nil,
	//			ExclusionRules:    nil,
	//		},
	//	},
	//},
	Conditions: []*nightfallAPI.Condition{
		{
			Detector: nightfallAPI.Detector{
				DetectorType:      nightfallAPI.DETECTORTYPE_NIGHTFALL_DETECTOR,
				NightfallDetector: nightfallAPI.NIGHTFALLDETECTORTYPE_API_KEY,
			},
		},
		{
			Detector: nightfallAPI.Detector{
				DetectorType:      nightfallAPI.DETECTORTYPE_NIGHTFALL_DETECTOR,
				NightfallDetector: nightfallAPI.NIGHTFALLDETECTORTYPE_CRYPTOGRAPHIC_KEY,
			},
		},
	},
	MaxNumberRoutines: DefaultMaxNumberRoutines,
}

// NightfallConfigFileStructure struct representation of nightfall config file
type NightfallConfigFileStructure struct {
	Conditions         []*nightfallAPI.Condition `json:"conditions"`
	MaxNumberRoutines  int                       `json:"maxNumberConcurrentRoutines"`
	TokenExclusionList []string                  `json:"tokenExclusionList"`
	FileInclusionList  []string                  `json:"fileInclusionList"`
	FileExclusionList  []string                  `json:"fileExclusionList"`
}

// Config general config struct
type Config struct {
	NightfallAPIKey            string
	NightfallConditions        []*nightfallAPI.Condition
	NightfallMaxNumberRoutines int
	TokenExclusionList         []string
	FileInclusionList          []string
	FileExclusionList          []string
}

// GetNightfallConfigFile loads nightfall config from file, returns default if missing/invalid
func GetNightfallConfigFile(workspacePath, fileName string, logger logger.Logger) (*NightfallConfigFileStructure, error) {
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
	var nightfallConfig NightfallConfigFileStructure
	err = json.Unmarshal(byteValue, &nightfallConfig)
	if err != nil {
		return nil, err
	}
	if len(nightfallConfig.Conditions) < 1 {
		return nil, errors.New("Nightfall config file is missing Conditions")
	}
	if nightfallConfig.MaxNumberRoutines <= 0 {
		nightfallConfig.MaxNumberRoutines = DefaultMaxNumberRoutines
	} else if nightfallConfig.MaxNumberRoutines > MaxConcurrentRoutinesCap {
		nightfallConfig.MaxNumberRoutines = MaxConcurrentRoutinesCap
	}
	nightfallConfig.FileExclusionList = append(nightfallConfig.FileExclusionList, nightfallConfigFilename)
	return &nightfallConfig, nil
}

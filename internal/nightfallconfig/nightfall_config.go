package nightfallconfig

import (
	"encoding/json"
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
const defaultDetectorsInfoMessage = "Using default detectors (API_KEY and CRYTOGRAPHIC_TOKEN)"

var apiKeyDetector = nightfallAPI.API_KEY
var cryptoKeyDetector = nightfallAPI.CRYPTOGRAPHIC_TOKEN

// NightfallConfigFileStructure struct representation of nightfall config file
type NightfallConfigFileStructure struct {
	Detectors          []*nightfallAPI.Detector `json:"detectors"`
	MaxNumberRoutines  int                      `json:"maxNumberConcurrentRoutines"`
	TokenExclusionList []string                 `json:"tokenExclusionList"`
	FileInclusionList  []string                 `json:"fileInclusionList"`
	FileExclusionList  []string                 `json:"fileExclusionList"`
}

// Config general config struct
type Config struct {
	NightfallAPIKey            string
	NightfallDetectors         []*nightfallAPI.Detector
	NightfallMaxNumberRoutines int
	TokenExclusionList         []string
	FileInclusionList          []string
	FileExclusionList          []string
}

// GetNightfallConfigFile loads nightfall config from file, returns default if missing/invalid
func GetNightfallConfigFile(workspacePath, fileName string, logger logger.Logger) *NightfallConfigFileStructure {
	defaultNightfallConfig := &NightfallConfigFileStructure{
		Detectors:         []*nightfallAPI.Detector{&apiKeyDetector, &cryptoKeyDetector},
		MaxNumberRoutines: DefaultMaxNumberRoutines,
	}
	nightfallConfigFile, err := os.Open(path.Join(workspacePath, fileName))
	if err != nil {
		logger.Warning(fmt.Sprintf("Error opening nightfall config: %s", err.Error()))
		logger.Info(defaultDetectorsInfoMessage)
		return defaultNightfallConfig
	}
	defer nightfallConfigFile.Close()
	byteValue, err := ioutil.ReadAll(nightfallConfigFile)
	if err != nil {
		logger.Warning(fmt.Sprintf("Error reading nightfall config: %s", err.Error()))
		logger.Info(defaultDetectorsInfoMessage)
		return defaultNightfallConfig
	}
	var nightfallConfig NightfallConfigFileStructure
	err = json.Unmarshal(byteValue, &nightfallConfig)
	if err != nil {
		logger.Warning("Invalid nightfall config")
		logger.Info(defaultDetectorsInfoMessage)
		return defaultNightfallConfig
	}
	if len(nightfallConfig.Detectors) < 1 {
		logger.Warning("Nightfall config requires at least one detector")
		logger.Info(defaultDetectorsInfoMessage)
		return defaultNightfallConfig
	}
	if nightfallConfig.MaxNumberRoutines > MaxConcurrentRoutinesCap || nightfallConfig.MaxNumberRoutines <= 0 {
		nightfallConfig.MaxNumberRoutines = DefaultMaxNumberRoutines
	}
	nightfallConfig.FileExclusionList = append(nightfallConfig.FileExclusionList, nightfallConfigFilename)
	return &nightfallConfig
}

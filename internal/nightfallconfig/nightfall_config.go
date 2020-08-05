package nightfallconfig

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"

	nightfallAPI "github.com/nightfallai/nightfall_go_client/generated"
)

// DetectorConfig struct containing all detectors being used and their
// associated minimum acceptable likelihood level
type DetectorConfig map[nightfallAPI.Detector]nightfallAPI.Likelihood

// NightfallConfigFileStructure struct representation of nightfall config file
type NightfallConfigFileStructure struct {
	Detectors          DetectorConfig `json:"detectors"`
	TokenExclusionList []string       `json:"tokenExclusionList"`
	FileInclusionList  []string       `json:"fileInclusionList"`
	FileExclusionList  []string       `json:"fileExclusionList"`
}

// Config general config struct
type Config struct {
	NightfallAPIKey    string
	NightfallDetectors DetectorConfig
	TokenExclusionList []string
	FileInclusionList  []string
	FileExclusionList  []string
}

// GetConfigFile loads nightfall config from file
func GetConfigFile(workspacePath, fileName string) (*NightfallConfigFileStructure, error) {
	nightfallConfigFile, err := os.Open(path.Join(workspacePath, fileName))
	if err != nil {
		return nil, err
	}
	defer nightfallConfigFile.Close()
	byteValue, err := ioutil.ReadAll(nightfallConfigFile)
	if err != nil {
		return nil, err
	}
	var nightfallConfig NightfallConfigFileStructure
	err = json.Unmarshal(byteValue, &nightfallConfig)
	if err != nil {
		return nil, err
	}
	if len(nightfallConfig.Detectors) < 1 {
		return nil, errors.New("Nightfall config file is missing detectors")
	}

	return &nightfallConfig, nil
}

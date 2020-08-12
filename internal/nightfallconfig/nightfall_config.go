package nightfallconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/davecgh/go-spew/spew"

	nightfallAPI "github.com/nightfallai/nightfall_go_client/generated"
)

const defaultMaxNumberRoutines = 30

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

// GetNightfallConfigFile loads nightfall config from file
func GetNightfallConfigFile(workspacePath, fileName string) (*NightfallConfigFileStructure, error) {
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
	fmt.Println("SPEWING DETECTORS")
	spew.Dump(nightfallConfig.Detectors)
	if len(nightfallConfig.Detectors) < 1 {
		return nil, errors.New("Nightfall config file is missing detectors")
	}
	if nightfallConfig.MaxNumberRoutines == 0 {
		nightfallConfig.MaxNumberRoutines = defaultMaxNumberRoutines
	}

	return &nightfallConfig, nil
}

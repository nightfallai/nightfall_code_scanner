package nightfall

import (
	"github.com/nightfallai/nightfall_cli/internal/interfaces/nightfallintf"
	nightfallAPI "github.com/nightfallai/nightfall_go_client/generated"
)

// APIClient is a wrapper around nightfallAPI.APIClient
type APIClient struct {
	APIClient *nightfallAPI.APIClient
}

// NewAPIClient generates a nightfallAPI client
func NewAPIClient() *APIClient {
	APIConfig := nightfallAPI.NewConfiguration()
	return &APIClient{
		APIClient: nightfallAPI.NewAPIClient(APIConfig),
	}
}

// ScanAPI gets the ScanApi from the api client
func (c *APIClient) ScanAPI() nightfallintf.NightfallScanAPI {
	return c.APIClient.ScanApi
}

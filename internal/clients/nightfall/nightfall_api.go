package nightfall

import (
	"context"
	"net/http"

	nightfallAPI "github.com/nightfallai/nightfall_go_client/generated"
)

// APIClient is a wrapper around nightfallAPI.APIClient
type APIClient struct {
	APIClient *nightfallAPI.APIClient
	//ScanApi   nightfallintf.NightfallScanAPI
}

//// NewAPIClient generates a nightfallAPI client
//func NewAPIClient(scanAPI nightfallintf.NightfallScanAPI) *APIClient {
//	APIConfig := nightfallAPI.NewConfiguration()
//	apiClient := nightfallAPI.NewAPIClient(APIConfig)
//	if scanAPI == nil {
//		scanAPI = apiClient.ScanApi
//	}
//	return &APIClient{
//		APIClient: nightfallAPI.NewAPIClient(APIConfig),
//		ScanApi:   scanAPI,
//	}
//}

//// ScanAPI gets the ScanApi from the api client
//func (c *APIClient) ScanAPI() nightfallintf.NightfallScanAPI {
//	return c.ScanApi
//}

// NewAPIClient generates a nightfallAPI client
func NewAPIClient() *APIClient {
	APIConfig := nightfallAPI.NewConfiguration()
	return &APIClient{
		APIClient: nightfallAPI.NewAPIClient(APIConfig),
	}
}

// ScanAPI gets the ScanApi from the api client
//func (c *APIClient) ScanAPI() nightfallintf.NightfallScanAPI {
//	return c.APIClient.ScanApi
//}

func (c *APIClient) ScanPayload(
	ctx context.Context,
	scanReq nightfallAPI.ScanRequest,
) ([][]nightfallAPI.ScanResponse, *http.Response, error) {
	return c.APIClient.ScanApi.ScanPayload(ctx, scanReq)
}

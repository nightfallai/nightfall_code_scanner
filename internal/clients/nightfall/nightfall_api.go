package nightfall

import (
	"context"
	"net/http"

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

// ScanPayload makes the scan request to the nightfallAPI
func (c *APIClient) ScanPayload(
	ctx context.Context,
	scanReq nightfallAPI.ScanRequestV2,
) ([][]nightfallAPI.ScanResponseV2, *http.Response, error) {
	return c.APIClient.ScanV2Api.ScanPayloadV2(ctx, scanReq)
}

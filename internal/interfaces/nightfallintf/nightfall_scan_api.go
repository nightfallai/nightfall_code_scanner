package nightfallintf

import (
	"context"
	"net/http"

	nightfallAPI "github.com/nightfallai/nightfall_go_client/generated"
)

//go:generate go run github.com/golang/mock/mockgen -destination=../../mocks/clients/nightfallscanapi_mock/nightfallscanapi_mock.go -source=../nightfallintf/nightfall_scan_api.go -package=nightfallscanapi_mock -mock_names=NightfallScanAPI=NightfallScanAPI

type NightfallScanAPI interface {
	ScanPayload(ctx context.Context, scanReq nightfallAPI.ScanRequest) ([][]nightfallAPI.ScanResponse, *http.Response, error)
}

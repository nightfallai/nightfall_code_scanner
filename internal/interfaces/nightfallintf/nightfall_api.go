package nightfallintf

import (
	"context"
	"net/http"

	nightfallAPI "github.com/nightfallai/nightfall_go_client/generated"
)

//go:generate go run github.com/golang/mock/mockgen -destination=../../mocks/clients/nightfallapi_mock/nightfallapi_mock.go -source=../nightfallintf/nightfall_api.go -package=nightfallapi_mock -mock_names=NightfallAPI=NightfallAPI

type NightfallAPI interface {
	//ScanAPI() NightfallScanAPI
	ScanPayload(ctx context.Context, scanReq nightfallAPI.ScanRequest) ([][]nightfallAPI.ScanResponse, *http.Response, error)
}

package nightfallintf

//go:generate go run github.com/golang/mock/mockgen -destination=../../mocks/clients/nightfallapi_mock/nightfallapi_mock.go -source=../nightfallintf/nightfall_api.go -package=nightfallapi_mock -mock_names=NightfallAPI=NightfallAPI

type NightfallAPI interface {
	ScanAPI() NightfallScanAPI
}

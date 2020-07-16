package interfaces

import (
	"net/http"
)

//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/clients/http_client.go -source=../interfaces/http_client.go -package=mock -mock_names=HTTPClient=HTTPClient

// HTTPClient is an interface for a http client
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

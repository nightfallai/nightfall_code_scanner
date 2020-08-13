package nightfall_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/nightfallai/nightfall_cli/internal/clients/diffreviewer"
	githublogger "github.com/nightfallai/nightfall_cli/internal/clients/logger/github_logger"
	"github.com/nightfallai/nightfall_cli/internal/clients/nightfall"
	"github.com/nightfallai/nightfall_cli/internal/mocks/clients/nightfallapi_mock"
	nightfallAPI "github.com/nightfallai/nightfall_go_client/generated"
	"github.com/stretchr/testify/suite"
)

const (
	exampleCreditCardNumber = "4916-6734-7572-5015"
	blurredCreditCard       = "49********"
	maxItemsForAPIReq       = 479
)

type nightfallTestSuite struct {
	suite.Suite
}

var expectedScanResponse = [][]nightfallAPI.ScanResponse{
	{},
	{
		{
			Fragment: exampleCreditCardNumber,
			Detector: string(nightfallAPI.CREDIT_CARD_NUMBER),
			Confidence: nightfallAPI.ScanResponseConfidence{
				Bucket: string(nightfallAPI.LIKELY),
			},
		},
	},
	{},
}

func (n *nightfallTestSuite) TestReviewDiff() {
	ctrl := gomock.NewController(n.T())
	defer ctrl.Finish()
	cc := nightfallAPI.CREDIT_CARD_NUMBER
	phone := nightfallAPI.PHONE_NUMBER
	detectors := []*nightfallAPI.Detector{&cc, &phone}
	mockAPIClient := nightfallapi_mock.NewNightfallAPI(ctrl)
	client := nightfall.Client{
		APIClient:         mockAPIClient,
		Detectors:         detectors,
		MaxNumberRoutines: 2,
	}

	numLines := 20
	numFiles := 50
	numScanReq := ((numLines * numFiles) + maxItemsForAPIReq - 1) / maxItemsForAPIReq
	filePath := "test/data"
	lineNum := 0
	content := fmt.Sprintf("this has a credit card number %s", exampleCreditCardNumber)

	lines := make([]*diffreviewer.Line, numLines)
	for i := range lines {
		lines[i] = &diffreviewer.Line{
			LnumNew: lineNum,
			Content: content,
		}
	}

	input := make([]*diffreviewer.FileDiff, numFiles)
	for i := range input {
		h := &diffreviewer.Hunk{
			Lines: lines,
		}
		input[i] = &diffreviewer.FileDiff{
			Hunks: []*diffreviewer.Hunk{
				h,
			},
			PathNew: filePath,
		}
	}

	c := diffreviewer.Comment{
		FilePath:   filePath,
		LineNumber: lineNum,
		Body:       fmt.Sprintf("Suspicious content detected (%s, type %s)", blurredCreditCard, nightfallAPI.CREDIT_CARD_NUMBER),
		Title:      fmt.Sprintf("Detected %s", nightfallAPI.CREDIT_CARD_NUMBER),
	}
	expectedComments := []*diffreviewer.Comment{&c, &c, &c}

	for i := 0; i < numScanReq; i++ {
		mockAPIClient.EXPECT().ScanPayload(gomock.Any(), gomock.AssignableToTypeOf(nightfallAPI.ScanRequest{})).Return(expectedScanResponse, nil, nil)
	}

	comments, err := client.ReviewDiff(context.Background(), githublogger.NewDefaultGithubLogger(), input)
	n.NoError(err, "Received error from ReviewDiff")
	n.Equal(expectedComments, comments, "Received incorrect response from ReviewDiff")
}

func (n *nightfallTestSuite) TestScan() {
	ctrl := gomock.NewController(n.T())
	defer ctrl.Finish()
	cc := nightfallAPI.CREDIT_CARD_NUMBER
	phone := nightfallAPI.PHONE_NUMBER
	detectors := []*nightfallAPI.Detector{&cc, &phone}
	mockAPIClient := nightfallapi_mock.NewNightfallAPI(ctrl)
	client := nightfall.Client{
		APIClient: mockAPIClient,
		Detectors: detectors,
	}

	items := []string{
		"this is a string",
		fmt.Sprintf("this has a credit card number %s", exampleCreditCardNumber),
		"tom cruise!!!!!!",
	}

	expectedScanReq := createScanReq(detectors, items)
	mockAPIClient.EXPECT().ScanPayload(gomock.Any(), gomock.AssignableToTypeOf(expectedScanReq)).Return(expectedScanResponse, nil, nil)

	resp, err := client.Scan(context.Background(), githublogger.NewDefaultGithubLogger(), items)
	n.NoError(err, "Received error from Scan")
	n.Equal(expectedScanResponse, resp, "Received incorrect response from Scan")
}

func (n *nightfallTestSuite) TestScanRetries() {
	ctrl := gomock.NewController(n.T())
	defer ctrl.Finish()
	cc := nightfallAPI.CREDIT_CARD_NUMBER
	phone := nightfallAPI.PHONE_NUMBER
	detectors := []*nightfallAPI.Detector{&cc, &phone}

	items := []string{
		"this is a string",
		fmt.Sprintf("this has a credit card number %s", exampleCreditCardNumber),
		"tom cruise!!!!!!",
	}

	expectedScanReq := createScanReq(detectors, items)
	expectedTooManyRequestsHTTPResponse := &http.Response{StatusCode: http.StatusTooManyRequests}

	tests := []struct {
		haveSuccess        bool
		haveNumReqAttempts int
		wantResponse       [][]nightfallAPI.ScanResponse
		desc               string
	}{
		{
			haveSuccess:        true,
			haveNumReqAttempts: 1,
			wantResponse:       expectedScanResponse,
			desc:               "success on first scan request",
		},
		{
			haveSuccess:        true,
			haveNumReqAttempts: 3,
			wantResponse:       expectedScanResponse,
			desc:               "success on third scan request attempt",
		},
		{
			haveSuccess:        false,
			haveNumReqAttempts: 5,
			wantResponse:       [][]nightfallAPI.ScanResponse(nil),
			desc:               "failed after max retries",
		},
	}
	for _, tt := range tests {
		mockAPIClient := nightfallapi_mock.NewNightfallAPI(ctrl)

		client := nightfall.Client{
			APIClient:         mockAPIClient,
			Detectors:         detectors,
			InitialRetryDelay: time.Millisecond,
		}

		numRetries := tt.haveNumReqAttempts
		if tt.haveSuccess {
			numRetries--
		}
		for i := 0; i < numRetries; i++ {
			mockAPIClient.EXPECT().ScanPayload(gomock.Any(), gomock.AssignableToTypeOf(expectedScanReq)).
				DoAndReturn(
					func(
						ctx context.Context,
						scanReq nightfallAPI.ScanRequest,
					) ([][]nightfallAPI.ScanResponse, *http.Response, error) {
						return [][]nightfallAPI.ScanResponse{nil}, expectedTooManyRequestsHTTPResponse, nightfall.ErrMaxScanRetries
					})
		}
		if tt.haveSuccess {
			mockAPIClient.EXPECT().ScanPayload(gomock.Any(), gomock.AssignableToTypeOf(expectedScanReq)).
				Return(expectedScanResponse, nil, nil)
		}
		resp, err := client.Scan(context.Background(), githublogger.NewDefaultGithubLogger(), items)
		if tt.haveSuccess {
			n.NoError(err, fmt.Sprintf("Received error from Scan in %s test", tt.desc))
		} else {
			n.EqualError(
				err,
				nightfall.ErrMaxScanRetries.Error(),
				fmt.Sprintf("Did not get error from %s test", tt.desc),
			)
		}
		n.Equal(tt.wantResponse, resp, fmt.Sprintf("Received incorrect response from Scan in %s test", tt.desc))
	}
}

func createScanReq(dets []*nightfallAPI.Detector, items []string) nightfallAPI.ScanRequest {
	detectors := make([]nightfallAPI.ScanRequestDetectors, 0, len(dets))
	for d := range dets {
		detectors = append(detectors, nightfallAPI.ScanRequestDetectors{
			Name: string(d),
		})
	}
	return nightfallAPI.ScanRequest{
		Detectors: detectors,
		Payload: nightfallAPI.ScanRequestPayload{
			Items: items,
		},
	}
}

func TestGithubClient(t *testing.T) {
	suite.Run(t, new(nightfallTestSuite))
}

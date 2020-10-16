package nightfall_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/golang/mock/gomock"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer"
	githublogger "github.com/nightfallai/nightfall_code_scanner/internal/clients/logger/github_logger"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/nightfall"
	"github.com/nightfallai/nightfall_code_scanner/internal/mocks/clients/nightfallapi_mock"
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

var testConditions = []*nightfallAPI.Condition{
	{
		Detector: nightfallAPI.Detector{
			DetectorType:      nightfallAPI.DETECTORTYPE_NIGHTFALL_DETECTOR,
			NightfallDetector: nightfallAPI.NIGHTFALLDETECTORTYPE_CREDIT_CARD_NUMBER,
		},
	},
	{
		Detector: nightfallAPI.Detector{
			DetectorType:      nightfallAPI.DETECTORTYPE_NIGHTFALL_DETECTOR,
			NightfallDetector: nightfallAPI.NIGHTFALLDETECTORTYPE_PHONE_NUMBER,
		},
	},
}
var testItems = []string{
	"this is a string",
	fmt.Sprintf("this has a credit card number %s", exampleCreditCardNumber),
	"tom cruise!!!!!!",
}

var expectedScanResponse = [][]nightfallAPI.ScanResponseV2{
	{},
	{
		{
			Fragment:     exampleCreditCardNumber,
			DetectorName: string(nightfallAPI.NIGHTFALLDETECTORTYPE_CREDIT_CARD_NUMBER),
			Confidence:   nightfallAPI.CONFIDENCE_LIKELY,
		},
	},
	{},
}

var expectedTooManyRequestsHTTPResponse = &http.Response{StatusCode: http.StatusTooManyRequests}
var expectedInternalServorErrorHTTPResponse = &http.Response{StatusCode: http.StatusInternalServerError}

func (n *nightfallTestSuite) TestReviewDiff() {
	ctrl := gomock.NewController(n.T())
	defer ctrl.Finish()
	mockAPIClient := nightfallapi_mock.NewNightfallAPI(ctrl)
	client := nightfall.Client{
		APIClient:         mockAPIClient,
		Conditions:        testConditions,
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
		Body:       fmt.Sprintf("Suspicious content detected (%s, type %s)", blurredCreditCard, nightfallAPI.NIGHTFALLDETECTORTYPE_CREDIT_CARD_NUMBER),
		Title:      fmt.Sprintf("Detected %s", nightfallAPI.NIGHTFALLDETECTORTYPE_CREDIT_CARD_NUMBER),
	}
	expectedComments := []*diffreviewer.Comment{&c, &c, &c}

	totalItems := make([]string, numLines*numFiles)
	for i := 0; i < numLines*numFiles; i++ {
		totalItems[i] = content
	}

	for i := 0; i < numScanReq; i++ {
		startIndex := i * maxItemsForAPIReq
		var endIndex int
		if len(totalItems) < startIndex+maxItemsForAPIReq {
			endIndex = len(totalItems)
		} else {
			endIndex = startIndex + maxItemsForAPIReq
		}

		expectedScanReq := client.CreateScanRequest(totalItems[startIndex:endIndex])
		mockAPIClient.EXPECT().ScanPayload(gomock.Any(), expectedScanReq).Return(expectedScanResponse, nil, nil)
	}

	comments, err := client.ReviewDiff(context.Background(), githublogger.NewDefaultGithubLogger(), input)
	n.NoError(err, "Received error from ReviewDiff")
	n.Equal(expectedComments, comments, "Received incorrect response from ReviewDiff")
}

func (n *nightfallTestSuite) TestReviewDiffConditionSetUUID() {
	ctrl := gomock.NewController(n.T())
	defer ctrl.Finish()
	mockAPIClient := nightfallapi_mock.NewNightfallAPI(ctrl)
	testConditionSetUUID := uuid.New().String()
	client := nightfall.Client{
		APIClient:         mockAPIClient,
		ConditionSetUUID:  testConditionSetUUID,
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
		Body:       fmt.Sprintf("Suspicious content detected (%s, type %s)", blurredCreditCard, nightfallAPI.NIGHTFALLDETECTORTYPE_CREDIT_CARD_NUMBER),
		Title:      fmt.Sprintf("Detected %s", nightfallAPI.NIGHTFALLDETECTORTYPE_CREDIT_CARD_NUMBER),
	}
	expectedComments := []*diffreviewer.Comment{&c, &c, &c}

	totalItems := make([]string, numLines*numFiles)
	for i := 0; i < numLines*numFiles; i++ {
		totalItems[i] = content
	}

	for i := 0; i < numScanReq; i++ {
		startIndex := i * maxItemsForAPIReq
		var endIndex int
		if len(totalItems) < startIndex+maxItemsForAPIReq {
			endIndex = len(totalItems)
		} else {
			endIndex = startIndex + maxItemsForAPIReq
		}

		expectedScanReq := client.CreateScanRequest(totalItems[startIndex:endIndex])
		mockAPIClient.EXPECT().ScanPayload(gomock.Any(), expectedScanReq).Return(expectedScanResponse, nil, nil)
	}

	comments, err := client.ReviewDiff(context.Background(), githublogger.NewDefaultGithubLogger(), input)
	n.NoError(err, "Received error from ReviewDiff")
	n.Equal(expectedComments, comments, "Received incorrect response from ReviewDiff")
}

func (n *nightfallTestSuite) TestSuccessfulScanPaths() {
	ctrl := gomock.NewController(n.T())
	defer ctrl.Finish()
	client := nightfall.Client{
		Conditions: testConditions,
	}
	expectedScanReq := client.CreateScanRequest(testItems)

	tests := []struct {
		haveNumRequests int
		wantResponse    [][]nightfallAPI.ScanResponseV2
		desc            string
	}{
		{
			haveNumRequests: 1,
			wantResponse:    expectedScanResponse,
			desc:            "success on first scan request",
		},
		{
			haveNumRequests: 3,
			wantResponse:    expectedScanResponse,
			desc:            "success on third scan request attempt",
		},
		{
			haveNumRequests: nightfall.MaxScanAttempts,
			wantResponse:    expectedScanResponse,
			desc:            "success on final scan request attempt",
		},
	}
	for _, tt := range tests {
		mockAPIClient := nightfallapi_mock.NewNightfallAPI(ctrl)
		client := nightfall.Client{
			APIClient:         mockAPIClient,
			Conditions:        testConditions,
			InitialRetryDelay: time.Millisecond,
		}

		mockAPIClient.EXPECT().ScanPayload(gomock.Any(), expectedScanReq).
			Return(
				[][]nightfallAPI.ScanResponseV2{nil},
				expectedTooManyRequestsHTTPResponse,
				nightfall.ErrMaxScanRetries,
			).Times(tt.haveNumRequests - 1)
		mockAPIClient.EXPECT().ScanPayload(gomock.Any(), expectedScanReq).
			Return(expectedScanResponse, nil, nil)

		resp, err := client.Scan(context.Background(), githublogger.NewDefaultGithubLogger(), testItems)
		n.NoError(err, fmt.Sprintf("Received unexpected error from Scan in %s test", tt.desc))
		n.Equal(
			tt.wantResponse,
			resp,
			fmt.Sprintf("Received incorrect response from Scan in %s test", tt.desc),
		)
	}
}

func (n *nightfallTestSuite) TestFailedScanPaths() {
	ctrl := gomock.NewController(n.T())
	defer ctrl.Finish()

	client := nightfall.Client{
		Conditions: testConditions,
	}
	expectedScanReq := client.CreateScanRequest(testItems)

	tests := []struct {
		haveNumRequests       int
		wantResponse          [][]nightfallAPI.ScanResponseV2
		wantFinalErr          error
		wantFinalHTTPResponse *http.Response
		desc                  string
	}{
		{
			haveNumRequests:       nightfall.MaxScanAttempts,
			wantResponse:          [][]nightfallAPI.ScanResponseV2(nil),
			wantFinalErr:          nightfall.ErrMaxScanRetries,
			wantFinalHTTPResponse: expectedTooManyRequestsHTTPResponse,
			desc:                  "failed after max retries",
		},
		{
			haveNumRequests:       3,
			wantResponse:          [][]nightfallAPI.ScanResponseV2(nil),
			wantFinalErr:          errors.New("500 Internal Server Error"),
			wantFinalHTTPResponse: expectedInternalServorErrorHTTPResponse,
			desc:                  "failed on third attempt with non-429 error",
		},
	}
	for _, tt := range tests {
		mockAPIClient := nightfallapi_mock.NewNightfallAPI(ctrl)
		client := nightfall.Client{
			APIClient:         mockAPIClient,
			Conditions:        testConditions,
			InitialRetryDelay: time.Millisecond,
		}

		mockAPIClient.EXPECT().ScanPayload(gomock.Any(), expectedScanReq).
			Return(
				[][]nightfallAPI.ScanResponseV2{nil},
				expectedTooManyRequestsHTTPResponse,
				nightfall.ErrMaxScanRetries,
			).Times(tt.haveNumRequests - 1)
		mockAPIClient.EXPECT().ScanPayload(gomock.Any(), expectedScanReq).
			Return([][]nightfallAPI.ScanResponseV2{nil}, tt.wantFinalHTTPResponse, tt.wantFinalErr)

		resp, err := client.Scan(context.Background(), githublogger.NewDefaultGithubLogger(), testItems)
		n.EqualError(
			err,
			tt.wantFinalErr.Error(),
			fmt.Sprintf("Did not get error from %s test", tt.desc),
		)
		n.Equal(tt.wantResponse, resp, fmt.Sprintf("Received incorrect response from Scan in %s test", tt.desc))
	}
}

func TestGithubClient(t *testing.T) {
	suite.Run(t, new(nightfallTestSuite))
}

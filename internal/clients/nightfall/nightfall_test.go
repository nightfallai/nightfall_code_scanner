package nightfall_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nightfallai/jenkins_test/internal/clients/diffreviewer"
	githublogger "github.com/nightfallai/jenkins_test/internal/clients/logger/github_logger"
	"github.com/nightfallai/jenkins_test/internal/clients/nightfall"
	"github.com/nightfallai/jenkins_test/internal/mocks/clients/nightfallapi_mock"
	"github.com/nightfallai/jenkins_test/internal/mocks/clients/nightfallscanapi_mock"
	"github.com/nightfallai/jenkins_test/internal/nightfallconfig"
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

func (n *nightfallTestSuite) TestReviewDiff() {
	ctrl := gomock.NewController(n.T())
	defer ctrl.Finish()
	detectorConfig := nightfallconfig.DetectorConfig{
		nightfallAPI.CREDIT_CARD_NUMBER: nightfallAPI.POSSIBLE,
		nightfallAPI.PHONE_NUMBER:       nightfallAPI.POSSIBLE,
	}
	mockAPIClient := nightfallapi_mock.NewNightfallAPI(ctrl)
	mockScanAPI := nightfallscanapi_mock.NewNightfallScanAPI(ctrl)
	client := nightfall.Client{
		APIClient:         mockAPIClient,
		DetectorConfigs:   detectorConfig,
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

	scanResp := [][]nightfallAPI.ScanResponse{
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
	c := diffreviewer.Comment{
		FilePath:   filePath,
		LineNumber: lineNum,
		Body:       fmt.Sprintf("Suspicious content detected (%s, type %s)", blurredCreditCard, nightfallAPI.CREDIT_CARD_NUMBER),
		Title:      fmt.Sprintf("Detected %s", nightfallAPI.CREDIT_CARD_NUMBER),
	}
	expectedComments := []*diffreviewer.Comment{&c, &c, &c}

	for i := 0; i < numScanReq; i++ {
		mockAPIClient.EXPECT().ScanAPI().Return(mockScanAPI)
		mockScanAPI.EXPECT().ScanPayload(gomock.Any(), gomock.Any()).Return(scanResp, nil, nil)
	}

	comments, err := client.ReviewDiff(context.Background(), githublogger.NewDefaultGithubLogger(), input)
	n.NoError(err, "Received error from ReviewDiff")
	n.Equal(expectedComments, comments, "Received incorrect response from ReviewDiff")
}

func (n *nightfallTestSuite) TestScan() {
	ctrl := gomock.NewController(n.T())
	defer ctrl.Finish()
	detectorConfig := nightfallconfig.DetectorConfig{
		nightfallAPI.CREDIT_CARD_NUMBER: nightfallAPI.POSSIBLE,
		nightfallAPI.PHONE_NUMBER:       nightfallAPI.POSSIBLE,
	}
	mockAPIClient := nightfallapi_mock.NewNightfallAPI(ctrl)
	mockScanAPI := nightfallscanapi_mock.NewNightfallScanAPI(ctrl)
	client := nightfall.Client{
		APIClient:       mockAPIClient,
		DetectorConfigs: detectorConfig,
	}

	items := []string{
		"this is a string",
		fmt.Sprintf("this has a credit card number %s", exampleCreditCardNumber),
		"tom cruise!!!!!!",
	}

	expectedScanReq := createScanReq(detectorConfig, items)
	expectedScanResp := [][]nightfallAPI.ScanResponse{
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

	mockAPIClient.EXPECT().ScanAPI().Return(mockScanAPI)
	mockScanAPI.EXPECT().ScanPayload(gomock.Any(), expectedScanReq).Return(expectedScanResp, nil, nil)

	resp, err := client.Scan(context.Background(), githublogger.NewDefaultGithubLogger(), items)
	n.NoError(err, "Received error from Scan")
	n.Equal(expectedScanResp, resp, "Received incorrect response from Scan")
}

func createScanReq(detectorConfig nightfallconfig.DetectorConfig, items []string) nightfallAPI.ScanRequest {
	detectors := make([]nightfallAPI.ScanRequestDetectors, 0, len(detectorConfig))
	for d := range detectorConfig {
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

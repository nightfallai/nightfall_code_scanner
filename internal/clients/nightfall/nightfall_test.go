package nightfall

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	nf "github.com/nightfallai/nightfall-go-sdk"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer"
	githublogger "github.com/nightfallai/nightfall_code_scanner/internal/clients/logger/github_logger"
	"github.com/stretchr/testify/assert"
)

var (
	blurredCreditCard = "49*****************"
	blurredAPIKey     = "yr**************************************"
)

type mockNightfall struct {
	scanFn func(context.Context, *nf.ScanTextRequest) (*nf.ScanTextResponse, error)
}

func (m *mockNightfall) ScanText(ctx context.Context, request *nf.ScanTextRequest) (*nf.ScanTextResponse, error) {
	return m.scanFn(ctx, request)
}

var testDetectionRules = []nf.DetectionRule{
	{
		Name: "my detection rule",
		Detectors: []nf.Detector{
			{
				MinNumFindings:    1,
				MinConfidence:     nf.ConfidencePossible,
				DisplayName:       "CRYPTOGRAPHIC_KEY",
				DetectorType:      nf.DetectorTypeNightfallDetector,
				NightfallDetector: "CRYPTOGRAPHIC_KEY",
			},
			{
				MinNumFindings:    1,
				MinConfidence:     nf.ConfidencePossible,
				DisplayName:       "API_KEY",
				DetectorType:      nf.DetectorTypeNightfallDetector,
				NightfallDetector: "API_KEY",
			},
		},
		LogicalOp: nf.LogicalOpAny,
	},
}
var testItems = []string{
	"this is a string",
	fmt.Sprintf("this has a credit card number %s", exampleCreditCardNumber),
	"tom cruise!!!!!!",
}

var expectedScanResponse = &nf.ScanTextResponse{
	Findings: [][]*nf.Finding{
		{},
		{
			{
				Finding:         exampleCreditCardNumber,
				RedactedFinding: "49*****************",
				Detector: nf.DetectorMetadata{
					DisplayName:  "CREDIT_CARD_NUMBER",
					DetectorUUID: "2136e3c9-feb0-4aea-8d3e-a767afabf501",
				},
				Confidence: string(nf.ConfidencePossible),
				Location: &nf.Location{ByteRange: &nf.Range{
					Start: 30,
					End:   40,
				}},
			},
		},
	},
}

func TestReviewDiff(t *testing.T) {
	mockAPIClient := &mockNightfall{}
	client := Client{
		APIClient:         mockAPIClient,
		DetectionRules:    testDetectionRules,
		MaxNumberRoutines: 1,
	}

	numLines := 20
	numFiles := 500
	numScanReq := ((numFiles) + maxItemsForAPIReq - 1) / maxItemsForAPIReq
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

	allLinesCombinedBuffer := bytes.NewBuffer(nil)
	for _, line := range lines {
		allLinesCombinedBuffer.WriteString(line.Content + " ")
	}
	allLinesCombined := allLinesCombinedBuffer.String()

	input := make([]*diffreviewer.FileDiff, numFiles)
	for i := range input {
		h := &diffreviewer.Hunk{
			Lines: lines,
		}
		input[i] = &diffreviewer.FileDiff{
			Hunks:   []*diffreviewer.Hunk{h},
			PathNew: filePath,
		}
	}

	c := diffreviewer.Comment{
		FilePath:   filePath,
		LineNumber: lineNum,
		Body:       fmt.Sprintf("Suspicious content detected (%q, type %q)", blurredCreditCard, "CREDIT_CARD_NUMBER"),
		Title:      fmt.Sprintf("Detected CREDIT_CARD_NUMBER"),
	}
	expectedComments := []*diffreviewer.Comment{&c, &c}

	totalItems := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		totalItems[i] = allLinesCombined
	}

	var callCount int
	expectedRequests := make([]*nf.ScanTextRequest, 0, numScanReq)
	for i := 0; i < numScanReq; i++ {
		startIndex := i * maxItemsForAPIReq
		var endIndex int
		if len(totalItems) < startIndex+maxItemsForAPIReq {
			endIndex = len(totalItems)
		} else {
			endIndex = startIndex + maxItemsForAPIReq
		}

		expectedScanReq := client.buildScanRequest(totalItems[startIndex:endIndex])
		expectedRequests = append(expectedRequests, expectedScanReq)
		mockAPIClient.scanFn = func(ctx context.Context, request *nf.ScanTextRequest) (*nf.ScanTextResponse, error) {
			assert.Equal(t, expectedRequests[callCount], request, "request object did not match")
			callCount++
			return expectedScanResponse, nil
		}
	}

	comments, err := client.ReviewDiff(context.Background(), githublogger.NewDefaultGithubLogger(), input)
	assert.NoError(t, err, "Received error from ReviewDiff")
	assert.Equal(t, expectedComments, comments, "Received incorrect response from ReviewDiff")
}

func TestReviewDiffDetectionRuleUUID(t *testing.T) {
	mockAPIClient := &mockNightfall{}
	client := Client{
		APIClient:          mockAPIClient,
		DetectionRuleUUIDs: []uuid.UUID{uuid.New()},
		MaxNumberRoutines:  1,
	}

	numLines := 20
	numFiles := 500
	numScanReq := (numFiles + maxItemsForAPIReq - 1) / maxItemsForAPIReq
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

	allLinesCombinedBuffer := bytes.NewBuffer(nil)
	for _, line := range lines {
		allLinesCombinedBuffer.WriteString(line.Content + " ")
	}
	allLinesCombined := allLinesCombinedBuffer.String()

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
		Body:       fmt.Sprintf("Suspicious content detected (%q, type %q)", blurredCreditCard, "CREDIT_CARD_NUMBER"),
		Title:      "Detected CREDIT_CARD_NUMBER",
	}
	expectedComments := []*diffreviewer.Comment{&c, &c}

	totalItems := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		totalItems[i] = allLinesCombined
	}

	var callCount int
	expectedRequests := make([]*nf.ScanTextRequest, 0, numScanReq)
	for i := 0; i < numScanReq; i++ {
		startIndex := i * maxItemsForAPIReq
		var endIndex int
		if len(totalItems) < startIndex+maxItemsForAPIReq {
			endIndex = len(totalItems)
		} else {
			endIndex = startIndex + maxItemsForAPIReq
		}

		expectedScanReq := client.buildScanRequest(totalItems[startIndex:endIndex])
		expectedRequests = append(expectedRequests, expectedScanReq)
		mockAPIClient.scanFn = func(ctx context.Context, request *nf.ScanTextRequest) (*nf.ScanTextResponse, error) {
			assert.Equal(t, expectedRequests[callCount], request, "request object did not match")
			callCount++
			return expectedScanResponse, nil
		}
	}

	comments, err := client.ReviewDiff(context.Background(), githublogger.NewDefaultGithubLogger(), input)
	assert.NoError(t, err, "Received error from ReviewDiff")
	assert.Equal(t, expectedComments, comments, "Received incorrect response from ReviewDiff")
}

func TestReviewDiffHasFindingMetadata(t *testing.T) {
	mockAPIClient := &mockNightfall{}
	client := Client{
		APIClient:         mockAPIClient,
		DetectionRules:    testDetectionRules,
		MaxNumberRoutines: 1,
	}

	numLines := 20
	numFiles := 500
	numScanReq := (numFiles + maxItemsForAPIReq - 1) / maxItemsForAPIReq
	filePath := "test/data"
	lineNum := 0
	content := fmt.Sprintf("this has a api key %s", exampleAPIKey)

	lines := make([]*diffreviewer.Line, numLines)
	for i := range lines {
		lines[i] = &diffreviewer.Line{
			LnumNew: lineNum,
			Content: content,
		}
	}

	allLinesCombinedBuffer := bytes.NewBuffer(nil)
	for _, line := range lines {
		allLinesCombinedBuffer.WriteString(line.Content + " ")
	}
	allLinesCombined := allLinesCombinedBuffer.String()

	input := make([]*diffreviewer.FileDiff, numFiles)
	for i := range input {
		h := &diffreviewer.Hunk{
			Lines: lines,
		}
		input[i] = &diffreviewer.FileDiff{
			Hunks:   []*diffreviewer.Hunk{h},
			PathNew: filePath,
		}
	}

	c := diffreviewer.Comment{
		FilePath:   filePath,
		LineNumber: lineNum,
		Body:       fmt.Sprintf("Suspicious content detected (%q, type %q (%s %s key))", blurredAPIKey, "API_KEY", "Active", "Stripe"),
		Title:      fmt.Sprintf("Detected API_KEY"),
	}
	expectedComments := []*diffreviewer.Comment{&c, &c}

	totalItems := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		totalItems[i] = allLinesCombined
	}

	scanResp := &nf.ScanTextResponse{
		Findings: [][]*nf.Finding{
			{},
			{
				{
					Finding:         exampleAPIKey,
					RedactedFinding: blurredAPIKey,
					Detector: nf.DetectorMetadata{
						DisplayName:  "API_KEY",
						DetectorUUID: "2136e3c9-feb0-4aea-8d3e-a767afabf501",
					},
					Confidence: string(nf.ConfidencePossible),
					FindingMetadata: &nf.FindingMetadata{APIKeyMetadata: &nf.APIKeyMetadata{
						Status: "ACTIVE",
						Kind:   "Stripe",
					}},
					Location: &nf.Location{ByteRange: &nf.Range{
						Start: 30,
						End:   40,
					}},
				},
			},
		},
	}

	var callCount int
	expectedRequests := make([]*nf.ScanTextRequest, numScanReq)
	// 500kb can only handle 426 files at a time, 500*1024/1200 = 426.66
	expectedRequests[0] = client.buildScanRequest(totalItems[0:426])
	expectedRequests[1] = client.buildScanRequest(totalItems[426:])
	for i := 0; i < numScanReq; i++ {
		mockAPIClient.scanFn = func(ctx context.Context, request *nf.ScanTextRequest) (*nf.ScanTextResponse, error) {
			assert.Equal(t, expectedRequests[callCount], request, "request object did not match")
			callCount++
			return scanResp, nil
		}
	}

	comments, err := client.ReviewDiff(context.Background(), githublogger.NewDefaultGithubLogger(), input)
	assert.NoError(t, err, "Received error from ReviewDiff")
	assert.Equal(t, expectedComments, comments, "Received incorrect response from ReviewDiff")
}

func TestScanPaths(t *testing.T) {
	client := Client{
		DetectionRules: testDetectionRules,
	}
	expectedScanReq := client.buildScanRequest(testItems)

	tests := []struct {
		haveNumRequests int
		wantResponse    *nf.ScanTextResponse
		wantErr         error
		desc            string
	}{
		{
			haveNumRequests: 1,
			wantResponse:    expectedScanResponse,
			desc:            "success on first scan request",
		},
		{
			haveNumRequests: maxScanAttempts + 1,
			wantResponse:    nil,
			wantErr:         errors.New("failure"),
			desc:            "failure from API",
		},
	}
	for _, tt := range tests {
		mockAPIClient := &mockNightfall{}
		client := Client{
			APIClient:         mockAPIClient,
			DetectionRules:    testDetectionRules,
			InitialRetryDelay: time.Millisecond,
		}

		mockAPIClient.scanFn = func(ctx context.Context, request *nf.ScanTextRequest) (*nf.ScanTextResponse, error) {
			assert.Equal(t, expectedScanReq, request, "request object did not match")
			if tt.haveNumRequests > maxScanAttempts {
				return nil, errors.New("failure")
			}
			return expectedScanResponse, nil
		}

		resp, err := client.Scan(context.Background(), testItems)
		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantResponse, resp)
	}
}

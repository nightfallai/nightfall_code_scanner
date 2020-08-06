package nightfall

import (
	"fmt"
	"testing"

	"github.com/nightfallai/jenkins_test/internal/clients/diffreviewer"
	"github.com/nightfallai/jenkins_test/internal/nightfallconfig"
	nightfallAPI "github.com/nightfallai/nightfall_go_client/generated"
	"github.com/stretchr/testify/assert"
)

const (
	exampleCreditCardNumber  = "4916-6734-7572-5015"
	exampleCreditCardNumber2 = "4242-4242-4242-4242"
	exampleCreditCardNumber3 = "4916-6734-7572-5016"
	exampleAPIKey            = "yr+ZWwIZp6ifFgaHV8410b2BxbRt5QiAj1EZx1qj"
	filePath                 = "file/path"
	lineNumber               = 0
	creditCardNumberContent  = "4916-6734-7572-5015 is my credit card number"
	creditCardNumber2Content = "this is my card num 4242-4242-4242-4242 - spend wisely"
	apiKeyContent            = "my api key is yr+ZWwIZp6ifFgaHV8410b2BxbRt5QiAj1EZx1qj"
)

var allLikelihoods = []nightfallAPI.Likelihood{
	nightfallAPI.VERY_UNLIKELY,
	nightfallAPI.UNLIKELY,
	nightfallAPI.POSSIBLE,
	nightfallAPI.LIKELY,
	nightfallAPI.VERY_LIKELY,
}

func TestChunkContent(t *testing.T) {
	tests := []struct {
		haveBufSize int
		haveLine    *diffreviewer.Line
		want        []*contentToScan
	}{
		{
			haveBufSize: 4,
			haveLine:    createLine("ABCDEFG"),
			want: []*contentToScan{
				createContentToScan("ABCD"),
				createContentToScan("EFG"),
			},
		},
		{
			haveBufSize: 6,
			haveLine:    createLine("ABCDEFG字"),
			want: []*contentToScan{
				createContentToScan("ABCDEF"),
				createContentToScan("G字"),
			},
		},
	}

	for _, tt := range tests {
		actual, err := chunkContent(tt.haveBufSize, tt.haveLine, filePath)
		assert.NoError(t, err, "Unexpected error from chunkContent")
		assert.Equal(t, tt.want, actual, "Incorrect response from chunkContent")
	}
}

func TestSliceListBySize(t *testing.T) {
	tests := []struct {
		haveIndex              int
		haveNumItemsForMaxSize int
		haveContentToScanList  []*contentToScan
		want                   []*contentToScan
	}{
		{
			haveIndex:              0,
			haveNumItemsForMaxSize: 2,
			haveContentToScanList: []*contentToScan{
				createContentToScan("1"),
				createContentToScan("2"),
				createContentToScan("3"),
				createContentToScan("4"),
				createContentToScan("5"),
			},
			want: []*contentToScan{
				createContentToScan("1"),
				createContentToScan("2"),
			},
		},
		{
			haveIndex:              4,
			haveNumItemsForMaxSize: 2,
			haveContentToScanList: []*contentToScan{
				createContentToScan("1"),
				createContentToScan("2"),
				createContentToScan("3"),
				createContentToScan("4"),
				createContentToScan("5"),
				createContentToScan("6"),
				createContentToScan("7"),
				createContentToScan("8"),
				createContentToScan("9"),
				createContentToScan("10"),
			},
			want: []*contentToScan{
				createContentToScan("9"),
				createContentToScan("10"),
			},
		},
		{
			haveIndex:              1,
			haveNumItemsForMaxSize: 3,
			haveContentToScanList: []*contentToScan{
				createContentToScan("1"),
				createContentToScan("2"),
				createContentToScan("3"),
				createContentToScan("4"),
			},
			want: []*contentToScan{
				createContentToScan("4"),
			},
		},
	}

	for _, tt := range tests {
		actual := sliceListBySize(tt.haveIndex, tt.haveNumItemsForMaxSize, tt.haveContentToScanList)
		assert.Equal(t, tt.want, actual, "Incorrect response from sliceListBySize")
	}
}

func TestCreateCommentsFromScanResp(t *testing.T) {
	detectorConfigs := nightfallconfig.DetectorConfig{
		nightfallAPI.CREDIT_CARD_NUMBER: nightfallAPI.LIKELY,
	}
	emptyTokenWhitelistMap := make(map[string]bool, 0)
	tokenWhitelistMap := map[string]bool{exampleCreditCardNumber2: true}
	creditCardResponse := createScanResponse(exampleCreditCardNumber, nightfallAPI.CREDIT_CARD_NUMBER, nightfallAPI.VERY_LIKELY)
	creditCard2Response := createScanResponse(exampleCreditCardNumber2, nightfallAPI.CREDIT_CARD_NUMBER, nightfallAPI.VERY_LIKELY)
	apiKeyResponse := createScanResponse(exampleAPIKey, nightfallAPI.API_KEY, nightfallAPI.VERY_LIKELY)
	tests := []struct {
		haveContentToScanList []*contentToScan
		haveScanResponseList  [][]nightfallAPI.ScanResponse
		haveTokenWhitelistMap map[string]bool
		want                  []*diffreviewer.Comment
		desc                  string
	}{
		{
			haveContentToScanList: []*contentToScan{
				createContentToScan(creditCardNumberContent),
				createContentToScan("nothing in here"),
				createContentToScan(apiKeyContent),
				createContentToScan(creditCardNumber2Content),
			},
			haveScanResponseList: [][]nightfallAPI.ScanResponse{
				{
					creditCardResponse,
				},
				{},
				{
					apiKeyResponse,
				},
				{
					creditCard2Response,
				},
			},
			haveTokenWhitelistMap: emptyTokenWhitelistMap,
			want: []*diffreviewer.Comment{
				createComment(creditCardResponse),
				createComment(creditCard2Response),
			},
			desc: "credit cards omit api finding",
		},
		{
			haveContentToScanList: []*contentToScan{
				createContentToScan(creditCardNumberContent),
				createContentToScan("nothing in here"),
				createContentToScan(apiKeyContent),
			},
			haveScanResponseList: [][]nightfallAPI.ScanResponse{
				{
					creditCardResponse,
				},
				{
					createScanResponse("low likelihood on 4534343", nightfallAPI.CREDIT_CARD_NUMBER, nightfallAPI.UNLIKELY),
				},
			},
			haveTokenWhitelistMap: emptyTokenWhitelistMap,
			want: []*diffreviewer.Comment{
				createComment(creditCardResponse),
			},
			desc: "single credit card passing likelihood threshold",
		},
		{
			haveContentToScanList: []*contentToScan{
				createContentToScan("nothing in here"),
				createContentToScan("nothing in here"),
				createContentToScan("nothing in here"),
				createContentToScan("nothing in here"),
				createContentToScan(apiKeyContent),
			},
			haveScanResponseList: [][]nightfallAPI.ScanResponse{
				{},
				{},
				{},
				{},
				{
					apiKeyResponse,
				},
			},
			haveTokenWhitelistMap: emptyTokenWhitelistMap,
			want:                  []*diffreviewer.Comment{},
			desc:                  "no comments",
		},
		{
			haveContentToScanList: []*contentToScan{
				createContentToScan(creditCardNumberContent),
				createContentToScan("nothing in here"),
				createContentToScan(apiKeyContent),
				createContentToScan(creditCardNumber2Content),
			},
			haveScanResponseList: [][]nightfallAPI.ScanResponse{
				{
					creditCardResponse,
				},
				{},
				{
					apiKeyResponse,
				},
				{
					creditCard2Response,
				},
			},
			haveTokenWhitelistMap: tokenWhitelistMap,
			want: []*diffreviewer.Comment{
				createComment(creditCardResponse),
			},
			desc: "single credit card whitelist",
		},
	}

	for _, tt := range tests {
		actual := createCommentsFromScanResp(tt.haveContentToScanList, tt.haveScanResponseList, detectorConfigs, tt.haveTokenWhitelistMap)
		assert.Equal(t, tt.want, actual, fmt.Sprintf("Incorrect response from createCommentsFromScanResp: %s test", tt.desc))
	}
}

func TestFoundSensitiveData(t *testing.T) {
	detectorConfigs := nightfallconfig.DetectorConfig{
		nightfallAPI.CREDIT_CARD_NUMBER: nightfallAPI.POSSIBLE,
	}
	tests := []struct {
		have nightfallAPI.Likelihood
		want bool
	}{}
	for _, l := range allLikelihoods {
		var want bool
		switch l {
		case nightfallAPI.VERY_UNLIKELY, nightfallAPI.UNLIKELY:
			want = false
		default:
			want = true
		}
		tests = append(tests, struct {
			have nightfallAPI.Likelihood
			want bool
		}{
			have: l,
			want: want,
		})
	}

	for _, tt := range tests {
		finding := createScanResponse("", nightfallAPI.CREDIT_CARD_NUMBER, tt.have)
		actual := foundSensitiveData(finding, detectorConfigs)
		assert.Equal(t, tt.want, actual, "Incorrect response from foundSensitiveData")
	}
}

func TestBlurContent(t *testing.T) {
	tests := []struct {
		have string
		want string
	}{
		{
			have: exampleCreditCardNumber,
			want: "49********",
		},
		{
			have: exampleAPIKey,
			want: "yr********",
		},
		{
			have: "汉字 Hello 123",
			want: "汉字********",
		},
		{
			have: "SHORT",
			want: "SH***",
		},
	}
	for _, tt := range tests {
		actual := blurContent(tt.have)
		assert.Equal(t, tt.want, actual, "Incorrect response from blurContent")
	}
}

func TestFilterFileDiffs(t *testing.T) {
	filePaths := []string{"path/secondary_path/file.txt", "a.go", "a/a.go", "test.go", "path/main.go", "path/test.py"}
	var fileDiffs []*diffreviewer.FileDiff
	for _, filePath := range filePaths {
		fileDiff := &diffreviewer.FileDiff{
			PathNew: filePath,
		}
		fileDiffs = append(fileDiffs, fileDiff)
	}

	tests := []struct {
		haveFileDiffs         []*diffreviewer.FileDiff
		haveInclusionFileList []string
		haveExclusionFileList []string
		wantFileDiffs         []*diffreviewer.FileDiff
		desc                  string
	}{
		{
			haveFileDiffs:         fileDiffs,
			haveInclusionFileList: nil,
			haveExclusionFileList: []string{},
			wantFileDiffs:         fileDiffs,
			desc:                  "empty inclusion and exclusion list",
		},
		{
			haveFileDiffs:         fileDiffs,
			haveInclusionFileList: []string{"path/*"},
			haveExclusionFileList: nil,
			wantFileDiffs:         []*diffreviewer.FileDiff{fileDiffs[0], fileDiffs[4], fileDiffs[5]},
			desc:                  "inclusion list only",
		},
		{
			haveFileDiffs:         fileDiffs,
			haveInclusionFileList: nil,
			haveExclusionFileList: []string{"test*"},
			wantFileDiffs:         []*diffreviewer.FileDiff{fileDiffs[0], fileDiffs[1], fileDiffs[2], fileDiffs[4]},
			desc:                  "exclusion list only",
		},
		{
			haveFileDiffs:         fileDiffs,
			haveInclusionFileList: []string{"\\.go$", "path*"},
			haveExclusionFileList: []string{"/secondary_path/*"},
			wantFileDiffs:         fileDiffs[1:6],
			desc:                  "inclusion and exclusion list",
		},
		{
			haveFileDiffs:         fileDiffs,
			haveInclusionFileList: []string{".*"},
			haveExclusionFileList: []string{},
			wantFileDiffs:         fileDiffs,
			desc:                  "include everything",
		},
		{
			haveFileDiffs:         fileDiffs,
			haveInclusionFileList: []string{".*"},
			haveExclusionFileList: []string{".*"},
			wantFileDiffs:         []*diffreviewer.FileDiff{},
			desc:                  "include and then exclude everything",
		},
	}
	for _, tt := range tests {
		actual := filterFileDiffs(tt.haveFileDiffs, tt.haveInclusionFileList, tt.haveExclusionFileList)
		assert.Equal(t, tt.wantFileDiffs, actual, fmt.Sprintf("Incorrect response from filter file diffs %s test", tt.desc))
	}
}

func createScanResponse(fragment string, detector nightfallAPI.Detector, likelihood nightfallAPI.Likelihood) nightfallAPI.ScanResponse {
	return nightfallAPI.ScanResponse{
		Fragment: fragment,
		Detector: string(detector),
		Confidence: nightfallAPI.ScanResponseConfidence{
			Bucket: string(likelihood),
		},
		Location: nightfallAPI.ScanResponseLocation{
			ByteRange: nightfallAPI.ScanResponseLocationByteRange{
				Start: 0,
				End:   int32(len(fragment)),
			},
			UnicodeRange: nightfallAPI.ScanResponseLocationUnicodeRange{
				Start: 0,
				End:   int32(len(fragment)),
			},
		},
	}
}

func createLine(content string) *diffreviewer.Line {
	return &diffreviewer.Line{
		Content: content,
		LnumNew: lineNumber,
	}
}

func createContentToScan(content string) *contentToScan {
	return &contentToScan{
		Content:    content,
		FilePath:   filePath,
		LineNumber: lineNumber,
	}
}

func createComment(finding nightfallAPI.ScanResponse) *diffreviewer.Comment {
	return &diffreviewer.Comment{
		Body:       getCommentMsg(finding),
		FilePath:   filePath,
		LineNumber: lineNumber,
		Title:      getCommentTitle(finding),
	}
}

package nightfall

import (
	"fmt"
	"testing"

	nf "github.com/nightfallai/nightfall-go-sdk"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer"
	githublogger "github.com/nightfallai/nightfall_code_scanner/internal/clients/logger/github_logger"
	"github.com/stretchr/testify/assert"
)

const (
	exampleCreditCardNumber  = "4916-6734-7572-5015"
	exampleCreditCardNumber2 = "4242-4242-4242-4242"
	exampleAPIKey            = "yr+ZWwIZp6ifFgaHV8410b2BxbRt5QiAj1EZx1qj"
	exampleIP                = "53.166.90.118"
	exampleLocalHostIP       = "127.0.0.1"
	filePath                 = "file/path"
	lineNumber               = 0
	creditCardNumberContent  = "4916-6734-7572-5015 is my credit card number"
	creditCardNumber2Content = "this is my card num 4242-4242-4242-4242 - spend wisely"
	apiKeyContent            = "my api key is yr+ZWwIZp6ifFgaHV8410b2BxbRt5QiAj1EZx1qj"
)

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
	emptyTokenExclusionList := make([]string, 0)
	creditCard2Regex := "4242-4242-4242-[0-9]{4}"
	localIpRegex := "^127\\."
	tokenExclusionList := []string{creditCard2Regex, localIpRegex}
	creditCardResponse := createFinding(exampleCreditCardNumber, "CREDIT_CARD_NUMBER")
	creditCard2Response := createFinding(exampleCreditCardNumber2, "CREDIT_CARD_NUMBER")
	apiKeyResponse := createFinding(exampleAPIKey, "API_KEY")
	ipAddressResponse := createFinding(exampleIP, "IP_ADDRESS")
	tests := []struct {
		haveContentToScanList  []*contentToScan
		haveScanResponse       nf.ScanTextResponse
		haveTokenExclusionList []string
		want                   []*diffreviewer.Comment
		desc                   string
	}{
		{
			haveContentToScanList: []*contentToScan{
				createContentToScan(creditCardNumberContent),
				createContentToScan("nothing in here"),
				createContentToScan(apiKeyContent),
				createContentToScan(creditCardNumber2Content),
			},
			haveScanResponse: nf.ScanTextResponse{
				Findings: [][]*nf.Finding{
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
			},
			haveTokenExclusionList: emptyTokenExclusionList,
			want: []*diffreviewer.Comment{
				createComment(creditCardResponse),
				createComment(apiKeyResponse),
				createComment(creditCard2Response),
			},
			desc: "credit cards and an api key",
		},
		{
			haveContentToScanList: []*contentToScan{
				createContentToScan("nothing in here"),
				createContentToScan("nothing in here"),
				createContentToScan("nothing in here"),
				createContentToScan("nothing in here"),
			},
			haveScanResponse: nf.ScanTextResponse{
				Findings: [][]*nf.Finding{
					{},
					{},
					{},
					{},
				},
			},
			haveTokenExclusionList: emptyTokenExclusionList,
			want:                   []*diffreviewer.Comment{},
			desc:                   "no comments",
		},
		{
			haveContentToScanList: []*contentToScan{
				createContentToScan("4242-4242-4242-abcd"),
				createContentToScan(exampleCreditCardNumber),
				createContentToScan(exampleCreditCardNumber2),
				createContentToScan(exampleLocalHostIP),
				createContentToScan(exampleIP),
			},
			haveScanResponse: nf.ScanTextResponse{
				Findings: [][]*nf.Finding{
					{},
					{
						creditCardResponse,
					},
					{},
					{},
					{
						ipAddressResponse,
					},
				},
			},
			haveTokenExclusionList: tokenExclusionList,
			want: []*diffreviewer.Comment{
				createComment(creditCardResponse),
				createComment(ipAddressResponse),
			},
			desc: "credit card and ip regex",
		},
	}
	for _, tt := range tests {
		actual := createCommentsFromScanResp(tt.haveContentToScanList, &tt.haveScanResponse, tt.haveTokenExclusionList)
		assert.Equal(t, tt.want, actual, fmt.Sprintf("Incorrect response from createCommentsFromScanResp: test '%s'", tt.desc))
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
	fileDiffs := make([]*diffreviewer.FileDiff, len(filePaths))
	for i, filePath := range filePaths {
		fileDiff := &diffreviewer.FileDiff{
			PathNew: filePath,
		}
		fileDiffs[i] = fileDiff
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
			haveExclusionFileList: []string{"*test*"},
			wantFileDiffs:         []*diffreviewer.FileDiff{fileDiffs[0], fileDiffs[1], fileDiffs[2], fileDiffs[4]},
			desc:                  "exclusion list only",
		},
		{
			haveFileDiffs:         fileDiffs,
			haveInclusionFileList: []string{"*.go", "path*"},
			haveExclusionFileList: []string{"*/secondary_path/*"},
			wantFileDiffs:         fileDiffs[1:6],
			desc:                  "inclusion and exclusion list",
		},
		{
			haveFileDiffs:         fileDiffs,
			haveInclusionFileList: []string{"*"},
			haveExclusionFileList: []string{},
			wantFileDiffs:         fileDiffs,
			desc:                  "include everything",
		},
		{
			haveFileDiffs:         fileDiffs,
			haveInclusionFileList: []string{"*"},
			haveExclusionFileList: []string{"*"},
			wantFileDiffs:         []*diffreviewer.FileDiff{},
			desc:                  "include and then exclude everything",
		},
	}
	for _, tt := range tests {
		actual := filterFileDiffs(tt.haveFileDiffs, tt.haveInclusionFileList, tt.haveExclusionFileList, githublogger.NewDefaultGithubLogger())
		assert.Equal(t, tt.wantFileDiffs, actual, fmt.Sprintf("Incorrect response from filter file diffs %s test", tt.desc))
	}
}

func TestMatchRegex(t *testing.T) {
	tests := []struct {
		haveStrs        []string
		havePatterns    []string
		wantMatchedStrs []string
		desc            string
	}{
		{
			haveStrs:        []string{"a.go", "b.py", "a/b/c.txt", "4242-4242-4242-4242"},
			havePatterns:    []string{".*"},
			wantMatchedStrs: []string{"a.go", "b.py", "a/b/c.txt", "4242-4242-4242-4242"},
			desc:            ".*",
		},
		{
			haveStrs:        []string{"301-123-4567", "1-240-925-5721", "7428501824", "127.253.42.0", "13.47.149.67"},
			havePatterns:    []string{"^(1-)?\\d{3}-\\d{3}-\\d{4}$", "^127\\."},
			wantMatchedStrs: []string{"301-123-4567", "1-240-925-5721", "127.253.42.0"},
			desc:            "phone number and local ip addresses",
		},
	}
	for _, tt := range tests {
		matchedStrs := make([]string, 0, len(tt.haveStrs))
		for _, s := range tt.haveStrs {
			if matchRegex(s, tt.havePatterns) {
				matchedStrs = append(matchedStrs, s)
			}
		}
		assert.Equal(t, tt.wantMatchedStrs, matchedStrs, fmt.Sprintf("Incorrect response from match regex %s test", tt.desc))
	}
}

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		haveFilePaths    []string
		havePatterns     []string
		wantMatchedPaths []string
		desc             string
	}{
		{
			haveFilePaths:    []string{"a.go", "b.py", "a/b/c.txt"},
			havePatterns:     []string{"*"},
			wantMatchedPaths: []string{"a.go", "b.py", "a/b/c.txt"},
			desc:             "*",
		},
		{
			haveFilePaths:    []string{"path/a.go", "path/secondary_path/c.py", "path/secondary_path/tertiary_path.txt"},
			havePatterns:     []string{"path/*"},
			wantMatchedPaths: []string{"path/a.go", "path/secondary_path/c.py", "path/secondary_path/tertiary_path.txt"},
			desc:             "*path/*",
		},
		{
			haveFilePaths:    []string{"secondary_path/a.go", "path/secondary_path/c.py", "path/secondary_path/tertiary_path.txt"},
			havePatterns:     []string{"*/secondary_path*"},
			wantMatchedPaths: []string{"path/secondary_path/c.py", "path/secondary_path/tertiary_path.txt"},
			desc:             "*/secondary_path*",
		},
	}
	for _, tt := range tests {
		matchedPaths := make([]string, 0, len(tt.haveFilePaths))
		globs := compileGlobs(tt.havePatterns, githublogger.NewDefaultGithubLogger())
		for _, s := range tt.haveFilePaths {
			if matchGlob(s, globs) {
				matchedPaths = append(matchedPaths, s)
			}
		}
		assert.Equal(t, tt.wantMatchedPaths, matchedPaths, fmt.Sprintf("Incorrect response from match glob %s test", tt.desc))
	}
}
func createFinding(fragment string, detectorName string) *nf.Finding {
	end := int64(len(fragment))
	return &nf.Finding{
		Finding: fragment,
		Detector: nf.DetectorMetadata{
			DisplayName: detectorName,
		},
		Confidence: string(nf.ConfidenceLikely),
		Location: &nf.Location{
			ByteRange: &nf.Range{
				Start: 0,
				End:   end,
			},
			CodepointRange: &nf.Range{
				Start: 0,
				End:   end,
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

func createComment(finding *nf.Finding) *diffreviewer.Comment {
	return &diffreviewer.Comment{
		Body:       getCommentMsg(finding),
		FilePath:   filePath,
		LineNumber: lineNumber,
		Title:      getCommentTitle(finding),
	}
}

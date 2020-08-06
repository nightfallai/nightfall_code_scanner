package nightfall

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"unicode/utf8"

	"github.com/nightfallai/jenkins_test/internal/clients/diffreviewer"
	"github.com/nightfallai/jenkins_test/internal/clients/logger"
	"github.com/nightfallai/jenkins_test/internal/nightfallconfig"
	nightfallAPI "github.com/nightfallai/nightfall_go_client/generated"
)

const (
	// max request size is 500KB, so set max to 490Kb for buffer room
	// maxSizeBytes = 490000
	// max list size imposed by Nightfall API
	// maxListSize = 50000
	contentChunkByteSize = 1024
	// max number of items that can be sent to Nightfall API at a time
	maxItemsForAPIReq = 479
)

// likelihoodThresholdMap gives each likelihood an integer value representation
// the integer value can be used to determine relative importance and can
// allow for likelihoods to be compared directly
// eg. VERY_LIKELY > LIKELY since likelihoodThresholdMap[VERY_LIKELY] > likelihoodThresholdMap[LIKELY]
var likelihoodThresholdMap = map[nightfallAPI.Likelihood]int{
	nightfallAPI.VERY_UNLIKELY: 1,
	nightfallAPI.UNLIKELY:      2,
	nightfallAPI.POSSIBLE:      3,
	nightfallAPI.LIKELY:        4,
	nightfallAPI.VERY_LIKELY:   5,
}

// Client client which uses Nightfall API
// to determine findings from input strings
type Client struct {
	APIClient         *nightfallAPI.APIClient
	APIKey            string
	DetectorConfigs   nightfallconfig.DetectorConfig
	TokenWhitelistMap map[string]bool
	FileInclusionList []string
	FileExclusionList []string
}

// NewClient create Client
func NewClient(config nightfallconfig.Config) *Client {
	APIConfig := nightfallAPI.NewConfiguration()
	tokenWhitelistMap := makeStrMap(config.TokenExclusionList)
	n := Client{
		APIClient:         nightfallAPI.NewAPIClient(APIConfig),
		APIKey:            config.NightfallAPIKey,
		DetectorConfigs:   config.NightfallDetectors,
		TokenWhitelistMap: tokenWhitelistMap,
		FileInclusionList: config.FileInclusionList,
		FileExclusionList: config.FileExclusionList,
	}
	return &n
}

type contentToScan struct {
	Content    string
	FilePath   string
	LineNumber int
}

func foundSensitiveData(finding nightfallAPI.ScanResponse, detectorConfigs nightfallconfig.DetectorConfig) bool {
	minimumLikelihoodForDetector, ok := detectorConfigs[nightfallAPI.Detector(finding.Detector)]
	if !ok {
		return false
	}
	findingLikelihood := nightfallAPI.Likelihood(finding.Confidence.Bucket)
	return likelihoodThresholdMap[findingLikelihood] >= likelihoodThresholdMap[minimumLikelihoodForDetector]
}

func blurContent(content string) string {
	contentRune := []rune(content)
	blurredContent := string(contentRune[:2])
	blurLength := 8
	if len(contentRune[2:]) < blurLength {
		blurLength = len(contentRune[2:])
	}
	for i := 0; i < blurLength; i++ {
		blurredContent += "*"
	}
	return blurredContent
}

func getCommentMsg(finding nightfallAPI.ScanResponse) string {
	blurredContent := blurContent(finding.Fragment)
	return fmt.Sprintf("Suspicious content detected (%s, type %s)", blurredContent, finding.Detector)
}

func getCommentTitle(finding nightfallAPI.ScanResponse) string {
	return fmt.Sprintf("Detected %s", finding.Detector)
}

// wordSplitter is of type bufio.SplitFunc (https://golang.org/pkg/bufio/#SplitFunc)
// this function is used to determine how to chunk the reader input into bufio.Scanner.
// This function will create chunks of input buffer size, but will not chunk in the middle of
// a word.
func wordSplitter(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF {
		return 0, nil, io.EOF
	}
	indexEndOfLastValidWord := len(data)
	// walk from back of input []byte to the front
	// the loop is looking for the index of the last
	// valid rune in the input []byte
	r, _ := utf8.DecodeLastRune(data)
	for r == utf8.RuneError {
		if indexEndOfLastValidWord > 1 {
			indexEndOfLastValidWord--
			r, _ = utf8.DecodeLastRune(data[:indexEndOfLastValidWord])
		} else {
			// multi-byte word does not fit in buffer
			// so request more data in buffer to complete word
			return 0, nil, nil
		}
	}
	numBytesRead := indexEndOfLastValidWord
	readChunk := data[:indexEndOfLastValidWord]
	return numBytesRead, readChunk, nil
}

func chunkContent(setBufSize int, line *diffreviewer.Line, filePath string) ([]*contentToScan, error) {
	chunkedContent := []*contentToScan{}
	r := bytes.NewReader([]byte(line.Content))
	s := bufio.NewScanner(r)
	s.Split(wordSplitter)
	buf := make([]byte, setBufSize)
	s.Buffer(buf, bufio.MaxScanTokenSize)
	for s.Scan() && s.Err() == nil {
		strChunk := s.Text()
		if len(strChunk) > 0 {
			cts := contentToScan{
				Content:    strChunk,
				FilePath:   filePath,
				LineNumber: line.LnumNew,
			}
			chunkedContent = append(chunkedContent, &cts)
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return chunkedContent, nil
}

func sliceListBySize(index, numItemsForMaxSize int, contentToScanList []*contentToScan) []*contentToScan {
	startIndex := index * numItemsForMaxSize
	if startIndex > len(contentToScanList) {
		startIndex = len(contentToScanList)
	}
	endIndex := (index + 1) * numItemsForMaxSize
	if endIndex > len(contentToScanList) {
		endIndex = len(contentToScanList)
	}
	return contentToScanList[startIndex:endIndex]
}

func createCommentsFromScanResp(
	inputContent []*contentToScan,
	resp [][]nightfallAPI.ScanResponse,
	detectorConfigs nightfallconfig.DetectorConfig,
	tokenWhitelist map[string]bool,
) []*diffreviewer.Comment {
	comments := []*diffreviewer.Comment{}
	for j, findingList := range resp {
		for _, finding := range findingList {
			if foundSensitiveData(finding, detectorConfigs) &&
				!isFindingInTokenWhitelist(finding.Fragment, tokenWhitelist) {
				// Found sensitive info
				// Create comment if fragment is not in whitelist
				correspondingContent := inputContent[j]
				findingMsg := getCommentMsg(finding)
				findingTitle := getCommentTitle(finding)
				c := diffreviewer.Comment{
					FilePath:   correspondingContent.FilePath,
					LineNumber: correspondingContent.LineNumber,
					Body:       findingMsg,
					Title:      findingTitle,
				}
				comments = append(comments, &c)
			}
		}
	}
	return comments
}

func isFindingInTokenWhitelist(fragment string, tokenWhitelist map[string]bool) bool {
	if tokenWhitelist == nil {
		return false
	}
	_, ok := tokenWhitelist[fragment]
	return ok
}

func (n *Client) createScanRequest(items []string) nightfallAPI.ScanRequest {
	detectors := make([]nightfallAPI.ScanRequestDetectors, 0, len(n.DetectorConfigs))
	for d := range n.DetectorConfigs {
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

// Scan send /scan request to Nightfall API and return findings
func (n *Client) Scan(ctx context.Context, logger logger.Logger, items []string) ([][]nightfallAPI.ScanResponse, error) {
	APIKey := nightfallAPI.APIKey{
		Key:    n.APIKey,
		Prefix: "",
	}
	newCtx := context.WithValue(ctx, nightfallAPI.ContextAPIKey, APIKey)
	request := n.createScanRequest(items)
	resp, _, err := n.APIClient.ScanApi.ScanPayload(newCtx, request)
	if err != nil {
		logger.Error(fmt.Sprintf("Error from Nightfall API, unable to successfully scan %d items", len(items)))
		return nil, err
	}
	return resp, nil
}

// ReviewDiff will take in a diff, chunk the contents of the diff
// and send the chunks to the Nightfall API to determine if it
// contains sensitive data
func (n *Client) ReviewDiff(
	ctx context.Context,
	logger logger.Logger,
	fileDiffs []*diffreviewer.FileDiff,
) ([]*diffreviewer.Comment, error) {
	fileDiffs = filterFileDiffs(fileDiffs, n.FileInclusionList, n.FileExclusionList)
	contentToScanList := make([]*contentToScan, 0, len(fileDiffs))
	// Chunk fileDiffs content and store chunk and its metadata
	for _, fd := range fileDiffs {
		for _, hunk := range fd.Hunks {
			for _, line := range hunk.Lines {
				chunkedContent, err := chunkContent(contentChunkByteSize, line, fd.PathNew)
				if err != nil {
					logger.Error("Error chunking git diff")
					return nil, err
				}
				contentToScanList = append(contentToScanList, chunkedContent...)
			}
		}
	}

	comments := []*diffreviewer.Comment{}
	// Integer round up division
	numRequestsRequired := (len(contentToScanList) + maxItemsForAPIReq - 1) / maxItemsForAPIReq
	logger.Debug(fmt.Sprintf("Sending %d requests to Nightfall API", numRequestsRequired))
	for i := 0; i < numRequestsRequired; i++ {
		// Use max number of items to determine content to send in request
		cts := sliceListBySize(i, maxItemsForAPIReq, contentToScanList)

		// Pull out content strings for request
		items := make([]string, len(cts))
		for i, item := range cts {
			items[i] = item.Content
		}

		// send API request
		resp, err := n.Scan(ctx, logger, items)
		if err != nil {
			logger.Debug(fmt.Sprintf("Error sending request number %d with %d items", i+1, len(items)))
			return nil, err
		}

		// Determine findings from response and create comments
		createdComments := createCommentsFromScanResp(cts, resp, n.DetectorConfigs, n.TokenWhitelistMap)
		comments = append(comments, createdComments...)
	}

	return comments, nil
}

func filterFileDiffs(fileDiffs []*diffreviewer.FileDiff, fileIncludeList, fileExcludeList []string) []*diffreviewer.FileDiff {
	filteredFileDiffs := fileDiffs
	if fileIncludeList != nil && len(fileIncludeList) > 0 {
		filteredFileDiffs = filterByFilePath(filteredFileDiffs, fileIncludeList, true)
	}
	if fileExcludeList != nil && len(fileExcludeList) > 0 {
		filteredFileDiffs = filterByFilePath(filteredFileDiffs, fileExcludeList, false)
	}
	return filteredFileDiffs
}

func filterByFilePath(fileDiffs []*diffreviewer.FileDiff, regexPatterns []string, include bool) []*diffreviewer.FileDiff {
	filteredFileDiffs := []*diffreviewer.FileDiff{}
	for _, fd := range fileDiffs {
		if (matchRegex(fd.PathNew, regexPatterns) && include) ||
			(!matchRegex(fd.PathNew, regexPatterns) && !include) {
			filteredFileDiffs = append(filteredFileDiffs, fd)
		}
	}
	return filteredFileDiffs
}

func matchRegex(filePath string, regexPatterns []string) bool {
	for _, pattern := range regexPatterns {
		if matched, _ := regexp.MatchString(pattern, filePath); matched {
			return true
		}
	}
	return false
}

func makeStrMap(strSlice []string) map[string]bool {
	if strSlice == nil {
		return nil
	}
	strMap := make(map[string]bool, len(strSlice))
	for _, v := range strSlice {
		strMap[v] = true
	}
	return strMap
}

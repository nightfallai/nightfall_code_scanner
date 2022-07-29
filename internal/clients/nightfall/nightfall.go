package nightfall

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gobwas/glob"
	"github.com/google/uuid"
	nf "github.com/nightfallai/nightfall-go-sdk"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer"
	"github.com/nightfallai/nightfall_code_scanner/internal/clients/logger"
	"github.com/nightfallai/nightfall_code_scanner/internal/nightfallconfig"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	contentChunkByteSize = 1024
	// max number of items that can be sent to Nightfall API at a time
	maxItemsForAPIReq = 479
	// timeout for the total time spent sending scan requests and receiving responses for a diff
	defaultTimeout = time.Minute * 20
	// maximum attempts to Nightfall API upon receiving 429 Too Many Requests before failing
	maxScanAttempts = 5
	// initial delay before re-attempting scan request
	initialDelay = time.Second
)

// Client uses the Nightfall API to scan text for findings
type Client struct {
	APIClient interface {
		ScanText(ctx context.Context, request *nf.ScanTextRequest) (*nf.ScanTextResponse, error)
	}
	DetectionRuleUUIDs     []uuid.UUID
	DetectionRules         []nf.DetectionRule
	MaxNumberRoutines      int
	InitialRetryDelay      time.Duration
	TokenExclusionList     []string
	FileInclusionList      []string
	FileExclusionList      []string
	DefaultRedactionConfig *nf.RedactionConfig
}

func NewClient(config nightfallconfig.Config) (*Client, error) {
	client, err := nf.NewClient(nf.OptionAPIKey(config.NightfallAPIKey))
	if err != nil {
		return nil, err
	}
	return &Client{
		APIClient:              client,
		DetectionRuleUUIDs:     config.NightfallDetectionRuleUUIDs,
		DetectionRules:         config.NightfallDetectionRules,
		MaxNumberRoutines:      config.NightfallMaxNumberRoutines,
		InitialRetryDelay:      initialDelay,
		TokenExclusionList:     config.TokenExclusionList,
		FileInclusionList:      config.FileInclusionList,
		FileExclusionList:      config.FileExclusionList,
		DefaultRedactionConfig: config.DefaultRedactionConfig,
	}, nil
}

type contentToScan struct {
	Content    string
	FilePath   string
	LineNumber int
}

func getCommentMsg(finding *nf.Finding) string {
	if finding.Finding == "" && finding.RedactedFinding == "" {
		return ""
	}

	content := finding.RedactedFinding
	if content == "" {
		content = finding.Finding
	}

	return fmt.Sprintf("Suspicious content detected (%q, type %s)", content, getDisplayType(finding))
}

func getCommentTitle(finding *nf.Finding) string {
	if finding.Detector.DisplayName == "" {
		return ""
	}
	return fmt.Sprintf("Detected %s", finding.Detector.DisplayName)
}

func getDisplayType(finding *nf.Finding) string {
	displayType := finding.Detector.DisplayName
	// Check if there is additional context to add to the display type
	if finding.FindingMetadata != nil && finding.FindingMetadata.APIKeyMetadata != nil {
		apiKeyMd := finding.FindingMetadata.APIKeyMetadata
		if kind := apiKeyMd.Kind; kind != "UNSPECIFIED" {
			titledKind := cases.Title(language.English).String(strings.ToLower(kind))
			if status := apiKeyMd.Status; status != "UNVERIFIED" {
				titledStatus := cases.Title(language.English).String(strings.ToLower(status))
				displayType = fmt.Sprintf("%q (%s %s key)", finding.Detector.DisplayName, titledStatus, titledKind)
			} else {
				titledConfidence := cases.Title(language.English).String(strings.ToLower(finding.Confidence))
				displayType = fmt.Sprintf("%q (%s %s key)", finding.Detector.DisplayName, titledConfidence, titledKind)
			}
		}
	}
	return displayType
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
	chunkedContent := make([]*contentToScan, 0)
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

func createCommentsFromScanResp(inputContent []*contentToScan, resp *nf.ScanTextResponse, tokenExclusionList []string) []*diffreviewer.Comment {
	comments := make([]*diffreviewer.Comment, 0)
	for j, findingList := range resp.Findings {
		for _, finding := range findingList {
			if finding.Finding != "" && !isFindingInTokenExclusionList(finding.Finding, tokenExclusionList) {
				// Found sensitive info
				// Create comment if fragment is not in exclusion set
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

func isFindingInTokenExclusionList(fragment string, tokenExclusionList []string) bool {
	if tokenExclusionList == nil {
		return false
	}
	return matchRegex(fragment, tokenExclusionList)
}

func matchRegex(finding string, regexPatterns []string) bool {
	for _, pattern := range regexPatterns {
		if matched, _ := regexp.MatchString(pattern, finding); matched {
			return true
		}
	}
	return false
}

func (n *Client) buildScanRequest(items []string) *nf.ScanTextRequest {
	ruleUUIDStrs := make([]string, 0, len(n.DetectionRuleUUIDs))
	for _, id := range n.DetectionRuleUUIDs {
		ruleUUIDStrs = append(ruleUUIDStrs, id.String())
	}

	return &nf.ScanTextRequest{
		Payload: items,
		Config: &nf.Config{
			DetectionRules:         n.DetectionRules,
			DetectionRuleUUIDs:     ruleUUIDStrs,
			DefaultRedactionConfig: n.DefaultRedactionConfig,
		},
	}
}

func (n *Client) scanContent(
	ctx context.Context,
	cts []*contentToScan,
	requestNum int,
	logger logger.Logger,
) ([]*diffreviewer.Comment, error) {
	// Pull out content strings for request
	items := make([]string, len(cts))
	for i, item := range cts {
		items[i] = item.Content
	}

	// send API request
	resp, err := n.Scan(ctx, items)
	if err != nil {
		logger.Debug(fmt.Sprintf("Error sending request number %d with %d items: %v", requestNum, len(items), err))
		return nil, err
	}

	// Determine findings from response and create comments
	createdComments := createCommentsFromScanResp(cts, resp, n.TokenExclusionList)
	logger.Info(fmt.Sprintf("Got %d annotations for request #%d", len(createdComments), requestNum))
	return createdComments, nil
}

func (n *Client) scanAllContent(
	ctx context.Context,
	logger logger.Logger,
	cts []*contentToScan,
	commentCh chan<- []*diffreviewer.Comment,
) {
	defer close(commentCh)
	blockingCh := make(chan struct{}, n.MaxNumberRoutines)
	var wg sync.WaitGroup

	// Integer round up division
	numRequestsRequired := (len(cts) + maxItemsForAPIReq - 1) / maxItemsForAPIReq

	logger.Info(fmt.Sprintf("Sending %d requests to Nightfall API", numRequestsRequired))
	for i := 0; i < numRequestsRequired; i++ {
		// Use max number of items to determine content to send in request
		contentSlice := sliceListBySize(i, maxItemsForAPIReq, cts)

		wg.Add(1)
		blockingCh <- struct{}{}
		go func(loopCount int, cts []*contentToScan) {
			defer wg.Done()
			if ctx.Err() != nil {
				return
			}

			c, err := n.scanContent(ctx, cts, loopCount+1, logger)
			if err != nil {
				logger.Error(fmt.Sprintf("Unable to scan %d content items", len(cts)))
			} else {
				commentCh <- c
			}
			<-blockingCh
		}(i, contentSlice)
	}
	wg.Wait()
}

func (n *Client) Scan(ctx context.Context, items []string) (*nf.ScanTextResponse, error) {
	request := n.buildScanRequest(items)
	return n.APIClient.ScanText(ctx, request)
}

// ReviewDiff will take in a diff, chunk the contents of the diff
// and send the chunks to the Nightfall API to determine if it
// contains sensitive data
func (n *Client) ReviewDiff(ctx context.Context, logger logger.Logger, fileDiffs []*diffreviewer.FileDiff) ([]*diffreviewer.Comment, error) {
	fileDiffs = filterFileDiffs(fileDiffs, n.FileInclusionList, n.FileExclusionList, logger)
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

	commentCh := make(chan []*diffreviewer.Comment)
	newCtx, cancel := context.WithDeadline(ctx, time.Now().Add(defaultTimeout))
	defer cancel()

	go n.scanAllContent(newCtx, logger, contentToScanList, commentCh)

	comments := make([]*diffreviewer.Comment, 0)
	for {
		select {
		case c, chOpen := <-commentCh:
			if !chOpen {
				return comments, nil
			}
			comments = append(comments, c...)
		case <-newCtx.Done():
			return nil, newCtx.Err()
		}
	}
}

func filterFileDiffs(fileDiffs []*diffreviewer.FileDiff, fileIncludeList, fileExcludeList []string, logger logger.Logger) []*diffreviewer.FileDiff {
	if len(fileIncludeList) > 0 {
		fileDiffs = filterByFilePath(fileDiffs, fileIncludeList, true, logger)
	}
	if len(fileExcludeList) > 0 {
		fileDiffs = filterByFilePath(fileDiffs, fileExcludeList, false, logger)
	}
	return fileDiffs
}

func filterByFilePath(fileDiffs []*diffreviewer.FileDiff, globPatterns []string, include bool, logger logger.Logger) []*diffreviewer.FileDiff {
	filteredFileDiffs := make([]*diffreviewer.FileDiff, 0, len(fileDiffs))
	globs := compileGlobs(globPatterns, logger)
	for _, fd := range fileDiffs {
		matched := matchGlob(fd.PathNew, globs)
		// if include (file inclusion), append if the filename matches a glob pattern
		// if !include (file exclusion), append if the filename does not match any pattern
		if (matched && include) || (!matched && !include) {
			filteredFileDiffs = append(filteredFileDiffs, fd)
		}
	}
	return filteredFileDiffs
}

func compileGlobs(globPatterns []string, logger logger.Logger) []glob.Glob {
	globs := make([]glob.Glob, 0, len(globPatterns))
	for _, pattern := range globPatterns {
		compiledGlob, err := glob.Compile(pattern)
		if err != nil {
			logger.Warning(fmt.Sprintf("Unable to compile glob pattern: %s", pattern))
			continue
		}
		globs = append(globs, compiledGlob)
	}
	return globs
}

func matchGlob(filePath string, globs []glob.Glob) bool {
	for _, g := range globs {
		if g.Match(filePath) {
			return true
		}
	}
	return false
}

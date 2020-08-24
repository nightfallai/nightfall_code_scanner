package diffutils

import (
	"strings"

	"github.com/nightfallai/nightfall_code_scanner/internal/clients/diffreviewer"
)

func FilterFileDiffs(fileDiffs []*diffreviewer.FileDiff) []*diffreviewer.FileDiff {
	if len(fileDiffs) == 0 {
		return fileDiffs
	}
	filteredFileDiffs := []*diffreviewer.FileDiff{}
	for _, fileDiff := range fileDiffs {
		fileDiff.Hunks = filterHunks(fileDiff.Hunks)
		if len(fileDiff.Hunks) > 0 {
			filteredFileDiffs = append(filteredFileDiffs, fileDiff)
		}
	}
	return filteredFileDiffs
}

func filterHunks(hunks []*diffreviewer.Hunk) []*diffreviewer.Hunk {
	filteredHunks := []*diffreviewer.Hunk{}
	for _, hunk := range hunks {
		hunk.Lines = filterLines(hunk.Lines)
		if len(hunk.Lines) > 0 {
			filteredHunks = append(filteredHunks, hunk)
		}
	}
	return filteredHunks
}

func filterLines(lines []*diffreviewer.Line) []*diffreviewer.Line {
	filteredLines := []*diffreviewer.Line{}
	for _, line := range lines {
		if line.Type == diffreviewer.LineAdded && !whitespaceOnlyLine(line) {
			filteredLines = append(filteredLines, line)
		}
	}
	return filteredLines
}

func whitespaceOnlyLine(line *diffreviewer.Line) bool {
	return strings.TrimSpace(line.Content) == ""
}

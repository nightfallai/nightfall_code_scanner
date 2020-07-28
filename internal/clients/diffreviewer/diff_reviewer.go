package diffreviewer

import (
	"github.com/watchtowerai/nightfall_dlp/internal/clients/logger"
	"github.com/watchtowerai/nightfall_dlp/internal/nightfallconfig"
)

//go:generate go run github.com/golang/mock/mockgen -destination=../../mocks/clients/diffreviewer_mock/diffreviewer_mock.go -source=../../clients/diffreviewer/diff_reviewer.go -package=diffreviewer_mock -mock_names=DiffReviewer=DiffReviewer

// DiffReviewer is the interface type for writing Nightfall DLP reviews/comments to a code repository commit/PR
type DiffReviewer interface {
	// LoadConfig retrieves the necessary config values from files or environment
	LoadConfig(nightfallConfigFileName string) (*nightfallconfig.Config, error)
	// GetDiff fetches the diff from the code repository and return a parsed array of FileDiffs
	GetDiff() ([]*FileDiff, error)
	// WriteComments posts the Nightfall DLP findings as comments/a review to the diff
	WriteComments(comments []*Comment) error
	// GetLogger gets the logger for the diff reviewer
	GetLogger() logger.Logger
}

// Comment holds the info required to write a comment to the code host
type Comment struct {
	Title      string
	Body       string
	FilePath   string
	LineNumber int
}

// FileDiff represents a unified diff for a single file.
//
// Example:
//   --- oldname	2009-10-11 15:12:20.000000000 -0700
//   +++ newname	2009-10-11 15:12:30.000000000 -0700
type FileDiff struct {
	// the old path of the file
	PathOld string
	// the new path of the file
	PathNew string

	// the old timestamp (empty if not present)
	TimeOld string
	// the new timestamp (empty if not present)
	TimeNew string

	Hunks []*Hunk

	// extended header lines (e.g., git's "new mode <mode>", "rename from <path>", index fb14f33..c19311b 100644, etc.)
	Extended []string

	// TODO: we may want `\ No newline at end of file` information for both the old and new file.
}

// Hunk represents change hunks that contain the line differences in the file.
//
// Example:
//   @@ -1,3 +1,4 @@ optional section heading
//    unchanged, contextual line
//   -deleted line
//   +added line
//   +added line
//    unchanged, contextual line
//
//  -1 -> the starting line number of the old file
//  3  -> the number of lines the change hunk applies to for the old file
//  +1 -> the starting line number of the new file
//  4  -> the number of lines the change hunk applies to for the new file
type Hunk struct {
	// the starting line number of the old file
	StartLineOld int
	// the number of lines the change hunk applies to for the old file
	LineLengthOld int

	// the starting line number of the new file
	StartLineNew int
	// the number of lines the change hunk applies to for the new file
	LineLengthNew int

	// optional section heading
	Section string

	// the body lines of the hunk
	Lines []*Line
}

// LineType represents the type of diff line.
type LineType int

const (
	// LineUnchanged represents unchanged, contextual line preceded by ' '
	LineUnchanged LineType = iota + 1
	// LineAdded represents added line preceded by '+'
	LineAdded
	// LineDeleted represents deleted line preceded by '-'
	LineDeleted
)

// Line represents a diff line.
type Line struct {
	// type of this line
	Type LineType
	// the line content without a preceded character (' ', '+', '-')
	Content string

	// the line in the file to a position in the diff.
	// the number of lines down from the first "@@" hunk header in the file.
	// e.g. The line just below the "@@" line is position 1, the next line is
	// position 2, and so on. The position in the file's diff continues to
	// increase through lines of whitespace and additional hunks until a new file
	// is reached. It's equivalent to the `position` field of input for comment
	// API of GitHub https://developer.github.com/v3/pulls/comments/#input
	LnumDiff int

	// the line number of the old file for LineUnchanged and LineDeleted
	// type. 0 for LineAdded type.
	LnumOld int

	// the line number of the new file for LineUnchanged and LineAdded type.
	// 0 for LineDeleted type.
	LnumNew int
}

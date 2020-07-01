package diffreviewer

// Comment holds the info required to write a comment to the code host
type Comment struct {
	Body       string
	FilePath   string
	LineNumber string
}

// DiffReviewer is the interface type for TBD
type DiffReviewer interface{}

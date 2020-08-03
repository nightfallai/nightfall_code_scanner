package flag

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

const (
	helpFlag        = "help"
	helpDescription = "Display details for Nightfall DLP"

	debugFlag        = "debug"
	debugShorthand   = "d"
	debugDescription = "Enable debug logs"
)

// Values contains all values parsed from command line flags
type Values struct {
	Debug bool
}

// Parse parses flags from command line
func Parse() (*Values, bool) {
	fs := pflag.NewFlagSet("all flags", pflag.ContinueOnError)

	values := Values{}
	var help bool

	fs.BoolVar(&help, helpFlag, false, helpDescription)
	fs.BoolVarP(&values.Debug, debugFlag, debugShorthand, false, debugDescription)

	err := fs.Parse(os.Args[1:])
	if err != nil || help {
		if err != nil {
			fmt.Fprint(os.Stderr, err, "\n")
		}
		fmt.Fprint(os.Stderr, "Usage: Nightfall DLP is used to scan content for sensitive information\n\n")
		fs.PrintDefaults()
		return nil, true
	}

	return &values, false
}

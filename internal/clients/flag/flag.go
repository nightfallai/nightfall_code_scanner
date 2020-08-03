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

func usage() {
	fmt.Fprint(os.Stderr, "Usage: Nightfall DLP is used to scan content for sensitive information\n\n")
	pflag.PrintDefaults()
	os.Exit(2)
}

// Parse parses flags from command line
func Parse() *Values {
	pflag.Usage = usage

	values := Values{}
	var help bool

	pflag.BoolVar(&help, helpFlag, false, helpDescription)
	pflag.BoolVarP(&values.Debug, debugFlag, debugShorthand, false, debugDescription)

	pflag.Parse()

	if help {
		pflag.Usage()
	}

	return &values
}

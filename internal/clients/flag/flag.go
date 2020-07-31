package flag

import "github.com/spf13/pflag"

const (
	debugFlag        = "debug"
	debugShorthand   = "d"
	debugDescription = "Enable debug logs"
)

// Values contains all values parsed from command line flags
type Values struct {
	Debug bool
}

// Parse parses flags from command line
func Parse() *Values {
	values := Values{}

	pflag.BoolVarP(&values.Debug, debugFlag, debugShorthand, false, debugDescription)

	pflag.Parse()

	return &values
}

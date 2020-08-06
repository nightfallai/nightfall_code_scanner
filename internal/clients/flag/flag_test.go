package flag_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nightfallai/jenkins_test/internal/clients/flag"
)

func TestParse(t *testing.T) {
	tests := []struct {
		desc       string
		have       []string
		wantValues *flag.Values
		wantDone   bool
	}{
		{
			desc: "Debug flag",
			have: []string{"--debug"},
			wantValues: &flag.Values{
				Debug: true,
			},
			wantDone: false,
		},
		{
			desc: "Debug shorthand flag",
			have: []string{"-d"},
			wantValues: &flag.Values{
				Debug: true,
			},
			wantDone: false,
		},
		{
			desc: "No flags",
			have: []string{},
			wantValues: &flag.Values{
				Debug: false,
			},
			wantDone: false,
		},
		{
			desc:       "Help flag",
			have:       []string{"--help"},
			wantValues: nil,
			wantDone:   true,
		},
		{
			desc:       "Invalid flag",
			have:       []string{"--flagdoesnotexist"},
			wantValues: nil,
			wantDone:   true,
		},
	}

	for _, tt := range tests {
		values, done := flag.Parse(tt.have)
		assert.Equal(t, tt.wantValues, values, fmt.Sprintf("Values returned are incorrect for test %s", tt.desc))
		assert.Equal(t, tt.wantDone, done, fmt.Sprintf("Done returned is incorrect for test %s", tt.desc))
	}
}

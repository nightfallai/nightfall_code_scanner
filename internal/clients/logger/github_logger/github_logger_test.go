package githublogger_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/nightfallai/nightfall_cli/internal/clients/logger"
	githublogger "github.com/nightfallai/nightfall_cli/internal/clients/logger/github_logger"
	"gotest.tools/assert"
)

const (
	debugPrefix   = "::debug::"
	warningPrefix = "::warning::"
	errorPrefix   = "::error::"
)

// Test case strings used by all tests
var tests = []string{"test", "汉字 Hello 123", "*** this has stuff"}

func setupTest() (logger.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	logger := log.New(os.Stdout, "", 0)
	logger.SetOutput(&buf)

	ghLogger := githublogger.NewGithubLogger(logger)
	return ghLogger, &buf
}

func TestDebug(t *testing.T) {
	ghLogger, buf := setupTest()

	for _, tt := range tests {
		buf.Reset()
		ghLogger.Debug(tt)
		assert.Equal(t, fmt.Sprintf("%s%s\n", debugPrefix, tt), buf.String())
	}
}

func TestWarning(t *testing.T) {
	ghLogger, buf := setupTest()

	for _, tt := range tests {
		buf.Reset()
		ghLogger.Warning(tt)
		assert.Equal(t, fmt.Sprintf("%s%s\n", warningPrefix, tt), buf.String())
	}
}

func TestError(t *testing.T) {
	ghLogger, buf := setupTest()

	for _, tt := range tests {
		buf.Reset()
		ghLogger.Error(tt)
		assert.Equal(t, fmt.Sprintf("%s%s\n", errorPrefix, tt), buf.String())
	}
}

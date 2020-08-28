package circlelogger_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/nightfallai/nightfall_code_scanner/internal/clients/logger"
	circlelogger "github.com/nightfallai/nightfall_code_scanner/internal/clients/logger/circle_logger"
	"gotest.tools/assert"
)

const (
	debugPrefix   = "[INFO]"
	warningPrefix = "[WARNING]"
	errorPrefix   = "[ERROR]"
)

// Test case strings used by all tests
var tests = []string{"test", "汉字 Hello 123", "*** this has stuff"}

func setupTest() (logger.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	logger := log.New(os.Stdout, "", 0)
	logger.SetOutput(&buf)

	ciLogger := circlelogger.NewCircleLogger(logger)
	return ciLogger, &buf
}

func TestDebug(t *testing.T) {
	ciLogger, buf := setupTest()

	for _, tt := range tests {
		buf.Reset()
		ciLogger.Debug(tt)
		assert.Equal(t, fmt.Sprintf("%s %s\n", debugPrefix, tt), buf.String())
	}
}

func TestWarning(t *testing.T) {
	ciLogger, buf := setupTest()

	for _, tt := range tests {
		buf.Reset()
		ciLogger.Warning(tt)
		assert.Equal(t, fmt.Sprintf("%s %s\n", warningPrefix, tt), buf.String())
	}
}

func TestError(t *testing.T) {
	ciLogger, buf := setupTest()

	for _, tt := range tests {
		buf.Reset()
		ciLogger.Error(tt)
		assert.Equal(t, fmt.Sprintf("%s %s\n", errorPrefix, tt), buf.String())
	}
}

package circlelogger

import (
	"log"
	"os"

	"github.com/nightfallai/nightfall_code_scanner/internal/clients/logger"
)

const (
	debugPrefix   = "[INFO]"
	warningPrefix = "[WARNING]"
	errorPrefix   = "[ERROR]"
)

// CircleLogger logger for CircleCI
type CircleLogger struct {
	log *log.Logger
}

// NewDefaultCircleLogger creates a CircleCI logger
// with the default log.Logger set
func NewDefaultCircleLogger() logger.Logger {
	return NewCircleLogger(log.New(os.Stdout, "", 0))
}

// NewCircleLogger creates a new CircleLogger
func NewCircleLogger(logger *log.Logger) logger.Logger {
	return &CircleLogger{
		log: logger,
	}
}

// Debug logs a debug message
func (l *CircleLogger) Debug(msg string) {
	l.log.Printf("%s %s\n", debugPrefix, msg)
}

// Warning logs a warning message
func (l *CircleLogger) Warning(msg string) {
	l.log.Printf("%s %s\n", warningPrefix, msg)
}

// Error logs a error message
func (l *CircleLogger) Error(msg string) {
	l.log.Printf("%s %s\n", errorPrefix, msg)
}

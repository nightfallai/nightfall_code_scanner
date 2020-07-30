package githublogger

import (
	"log"
	"os"

	"github.com/nightfallai/jenkins_test/internal/clients/flag"
	"github.com/nightfallai/jenkins_test/internal/clients/logger"
)

const (
	debugPrefix   = "::debug::"
	warningPrefix = "::warning::"
	errorPrefix   = "::error::"
)

// GithubLogger logger for Github Actions
type GithubLogger struct {
	log   *log.Logger
	debug bool
}

// NewDefaultGithubLogger creates a github logger
// with the default log.Logger set
func NewDefaultGithubLogger(flagValues *flag.Values) logger.Logger {
	return NewGithubLogger(log.New(os.Stdout, "", 0), flagValues)
}

// NewGithubLogger creates a new GithubLogger
func NewGithubLogger(logger *log.Logger, flagValues *flag.Values) logger.Logger {
	return &GithubLogger{
		log:   logger,
		debug: flagValues.Debug,
	}
}

// Debug logs a debug message
// to view debug logs the Github secret
// ACTIONS_RUNNER_DEBUG must be set to true
// https://docs.github.com/en/actions/configuring-and-managing-workflows/managing-a-workflow-run#enabling-debug-logging
func (l *GithubLogger) Debug(msg string) {
	l.log.Printf("%s%s\n", debugPrefix, msg)
}

// Warning logs a warning message
func (l *GithubLogger) Warning(msg string) {
	l.log.Printf("%s%s\n", warningPrefix, msg)
}

// Error logs a error message
func (l *GithubLogger) Error(msg string) {
	l.log.Printf("%s%s\n", errorPrefix, msg)
}

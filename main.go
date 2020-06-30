package main

import (
	"github.com/watchtowerai/diff_reviewer/internal/clients"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// main starts the service process.
func main() {
	app := fx.New(
		fx.Provide(
			// general constructors
			newSugarLogger,
		),
		clients.Module,
	)
	app.Run()
}

// newSugarLogger helper
func newSugarLogger(zapLog *zap.Logger) *zap.SugaredLogger {
	return zapLog.Sugar()
}

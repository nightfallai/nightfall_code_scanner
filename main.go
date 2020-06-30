package main

import (
	"path"
	"runtime"

	"github.com/watchtowerai/watchtower_go_libraries/pkg/clients/datadog"
	"github.com/watchtowerai/watchtower_go_libraries/pkg/config"

	"github.com/spf13/viper"
	"github.com/watchtowerai/diff_reviewer/internal/clients"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// main starts the service process.
func main() {
	app := fx.New(
		fx.Provide(
			// general constructors
			newConfig,
			newSugarLogger,
			newRecovery,
		),
		clients.Module,
		fx.Invoke(
			datadog.NewTracer,
		),
	)
	app.Run()
}

func newConfig() (*viper.Viper, error) {
	return config.NewConfig(path.Join(getDirName(), "../config"))
}

// newSugarLogger helper
func newSugarLogger(zapLog *zap.Logger) *zap.SugaredLogger {
	return zapLog.Sugar()
}

func getDirName() string {
	_, dirname, _, ok := runtime.Caller(0)
	if !ok {
		panic("Cannot get working directory")
	}
	return dirname
}

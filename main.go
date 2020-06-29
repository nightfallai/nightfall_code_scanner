package main

import (
	"path"
	"runtime"

	"github.com/spf13/viper"
	"github.com/watchtowerai/diff_reviewer/internal/clients"
	"github.com/watchtowerai/watchtower_go_libraries/pkg/clients/bugsnag"
	"github.com/watchtowerai/watchtower_go_libraries/pkg/clients/datadog"
	"github.com/watchtowerai/watchtower_go_libraries/pkg/config"
	log "github.com/watchtowerai/watchtower_go_libraries/pkg/log"
	"github.com/watchtowerai/watchtower_go_libraries/pkg/modules/logging"
	"github.com/watchtowerai/watchtower_go_libraries/pkg/modules/recovery"
	recovery2 "github.com/watchtowerai/watchtower_go_libraries/pkg/recovery"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// main starts the service process.
func main() {
	app := fx.New(
		fx.Provide(
			// general constructors
			newConfig,
			logging.NewZapLoggerWithBugsnag,
			log.NewSugaredLoggerFactory,
			bugsnag.NewConfiguration,
			bugsnag.NewBugsnagNotifier,
			recovery.NewBugsnagHandler,
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

func newRecovery(bsh *recovery.BugsnagHandler) recovery2.Handler {
	return bsh
}

func getDirName() string {
	_, dirname, _, ok := runtime.Caller(0)
	if !ok {
		panic("Cannot get working directory")
	}
	return dirname
}

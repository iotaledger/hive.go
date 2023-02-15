package shutdown

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/hive.go/app/shutdown"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:      "Shutdown",
			Provide:   provide,
			DepsFunc:  func(cDeps dependencies) { deps = cDeps },
			Params:    params,
			Configure: configure,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
	deps          dependencies
)

type dependencies struct {
	dig.In
	ShutdownHandler *shutdown.ShutdownHandler
}

func provide(c *dig.Container) error {

	if err := c.Provide(func() *shutdown.ShutdownHandler {
		return shutdown.NewShutdownHandler(
			CoreComponent.Logger(),
			CoreComponent.Daemon(),
			shutdown.WithStopGracePeriod(ParamsShutdown.StopGracePeriod),
			shutdown.WithSelfShutdownLogsEnabled(ParamsShutdown.Log.Enabled),
			shutdown.WithSelfShutdownLogsFilePath(ParamsShutdown.Log.FilePath),
		)
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}

func configure() error {
	return deps.ShutdownHandler.Run()
}

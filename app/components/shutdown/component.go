package shutdown

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/hive.go/app/shutdown"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:    "Shutdown",
			Provide: provide,
			Params:  params,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
)

func provide(c *dig.Container) error {

	if err := c.Provide(func() (*shutdown.ShutdownHandler, error) {
		handler := shutdown.NewShutdownHandler(
			CoreComponent.Logger(),
			CoreComponent.Daemon(),
			shutdown.WithStopGracePeriod(ParamsShutdown.StopGracePeriod),
			shutdown.WithSelfShutdownLogsEnabled(ParamsShutdown.Log.Enabled),
			shutdown.WithSelfShutdownLogsFilePath(ParamsShutdown.Log.FilePath),
		)

		// start the handler to be able to catch shutdown signals during the provide stage
		if err := handler.Run(); err != nil {
			return nil, err
		}

		return handler, nil
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}

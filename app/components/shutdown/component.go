package shutdown

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/hive.go/app/shutdown"
)

func init() {
	Component = &app.Component{
		Name:    "Shutdown",
		Provide: provide,
		Params:  params,
	}
}

var (
	Component *app.Component
)

func provide(c *dig.Container) error {
	// we need to initialize and run the handler outside of the "Provide" function,
	// otherwise the shutdown handler might not be initialized if
	// it was not a dependency in another component.
	handler := shutdown.NewShutdownHandler(
		Component.Logger,
		Component.Daemon(),
		shutdown.WithStopGracePeriod(ParamsShutdown.StopGracePeriod),
		shutdown.WithSelfShutdownLogsEnabled(ParamsShutdown.Log.Enabled),
		shutdown.WithSelfShutdownLogsFilePath(ParamsShutdown.Log.FilePath),
	)

	// start the handler to be able to catch shutdown signals during the provide stage
	if err := handler.Run(); err != nil {
		Component.LogPanic(err.Error())
	}

	if err := c.Provide(func() *shutdown.ShutdownHandler {
		return handler
	}); err != nil {
		Component.LogPanic(err.Error())
	}

	return nil
}

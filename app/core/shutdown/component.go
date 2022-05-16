package shutdown

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
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
	ShutdownHandler *ShutdownHandler
}

func provide(c *dig.Container) error {

	if err := c.Provide(func() *ShutdownHandler {
		return NewShutdownHandler(CoreComponent.Logger(), CoreComponent.Daemon())
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}

func configure() error {
	deps.ShutdownHandler.Run()
	return nil
}

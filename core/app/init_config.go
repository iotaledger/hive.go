package app

import (
	"go.uber.org/dig"
)

// InitConfigParsFunc gets called with a dig.Container.
type InitConfigParsFunc func(c *dig.Container) error

// PreProvideFunc gets called with a dig.Container, the configs the InitComponent brings to the app and the InitConfig.
type PreProvideFunc func(c *dig.Container, application *App, initConf *InitConfig) error

// ProvideFunc gets called with a dig.Container.
type ProvideFunc func(c *dig.Container) error

// IsEnabledFunc gets called to check whether the Plugin is enabled.
// It returns true if the Plugin is enabled.
type IsEnabledFunc func() bool

// InitFunc gets called as the initialization function of the app.
type InitFunc func(application *App) error

// Callback is a function called without any arguments.
type Callback func() error

// InitConfig describes the result of a app initialization.
type InitConfig struct {
	forceDisabledComponents []string
}

// ForceDisableComponent is used to force disable components before the provide stage.
func (ic *InitConfig) ForceDisableComponent(identifier string) {
	exists := false
	for _, entry := range ic.forceDisabledComponents {
		if entry == identifier {
			exists = true

			break
		}
	}

	if !exists {
		ic.forceDisabledComponents = append(ic.forceDisabledComponents, identifier)
	}
}

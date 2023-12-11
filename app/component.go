package app

import (
	"os"
	"strings"
	"sync"

	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app/daemon"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/hive.go/runtime/workerpool"
)

// InitFunc gets called as the initialization function of the app.
type InitFunc func(application *App) error

// InitConfigParamsFunc gets called with a dig.Container.
type InitConfigParamsFunc func(c *dig.Container) error

// IsEnabledFunc gets called to check whether the Compoment is enabled.
// It gets called with a dig.Container and returns true if the Compoment is enabled.
type IsEnabledFunc func(c *dig.Container) bool

// ProvideFunc gets called with a dig.Container.
type ProvideFunc func(c *dig.Container) error

// Callback is a function called without any arguments.
type Callback func() error

// ComponentParams defines the parameters configuration of a component.
type ComponentParams struct {
	// Handler to add configuration parameters to the default config.
	Params map[string]any
	// Handler to add configuration parameters to the additional configs.
	AdditionalParams map[string]map[string]any
	// The configuration values to mask.
	Masked []string
}

// Component is something which extends the App's capabilities.
type Component struct {
	// A reference to the App instance.
	app *App
	// The name of the component.
	Name string
	// The config parameters for this component.
	Params *ComponentParams
	// The function to call to initialize the component dependencies.
	DepsFunc interface{}
	// InitConfigParams gets called in the init stage of app initialization.
	// This can be used to provide config parameters even if the component is disabled.
	InitConfigParams InitConfigParamsFunc
	// IsEnabled gets called to check whether the component is enabled.
	IsEnabled IsEnabledFunc
	// Provide gets called in the provide stage of app initialization (enabled components only).
	Provide ProvideFunc
	// Configure gets called in the configure stage of app initialization (enabled components only).
	Configure Callback
	// Run gets called in the run stage of app initialization (enabled components only).
	Run Callback
	// WorkerPool gets configured and started automatically for each component (enabled components only).
	WorkerPool *workerpool.WorkerPool

	// The logger instance used in this component.
	logger     log.Logger
	loggerOnce sync.Once
}

func (c *Component) App() *App {
	return c.app
}

// Logger instantiates and returns a logger with the name of the component.
func (c *Component) Logger() log.Logger {
	c.loggerOnce.Do(func() {
		c.logger, _ = c.App().NewLogger(c.Name)
	})

	return c.logger
}

func (c *Component) Daemon() daemon.Daemon {
	return c.App().Daemon()
}

func (c *Component) Identifier() string {
	return strings.ToLower(strings.ReplaceAll(c.Name, " ", ""))
}

// LogDebug uses fmt.Sprint to construct and log a message.
func (c *Component) LogDebug(msg string, args ...interface{}) {
	c.Logger().LogDebug(msg, args...)
}

// LogDebugf uses fmt.Sprintf to log a templated message.
func (c *Component) LogDebugf(template string, args ...interface{}) {
	c.Logger().LogDebugf(template, args...)
}

// LogError uses fmt.Sprint to construct and log a message.
func (c *Component) LogError(msg string, args ...interface{}) {
	c.Logger().LogError(msg, args...)
}

// LogErrorAndExit uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (c *Component) LogErrorAndExit(msg string, args ...interface{}) {
	c.Logger().LogError(msg, args...)
	c.Logger().LogError("Exiting...")
	os.Exit(1)
}

// LogErrorf uses fmt.Sprintf to log a templated message.
func (c *Component) LogErrorf(template string, args ...interface{}) {
	c.Logger().LogErrorf(template, args...)
}

// LogErrorfAndExit uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (c *Component) LogErrorfAndExit(template string, args ...interface{}) {
	c.Logger().LogErrorf(template, args...)
	c.Logger().LogError("Exiting...")
	os.Exit(1)
}

// LogInfo uses fmt.Sprint to construct and log a message.
func (c *Component) LogInfo(msg string, args ...interface{}) {
	c.Logger().LogInfo(msg, args...)
}

// LogInfof uses fmt.Sprintf to log a templated message.
func (c *Component) LogInfof(template string, args ...interface{}) {
	c.Logger().LogInfof(template, args...)
}

// LogWarn uses fmt.Sprint to construct and log a message.
func (c *Component) LogWarn(msg string, args ...interface{}) {
	c.Logger().LogWarn(msg, args...)
}

// LogWarnf uses fmt.Sprintf to log a templated message.
func (c *Component) LogWarnf(template string, args ...interface{}) {
	c.Logger().LogWarnf(template, args...)
}

// LogFatal uses fmt.Sprint to construct and log a message.
func (c *Component) LogFatal(msg string, args ...interface{}) {
	c.Logger().LogFatal(msg, args...)
}

// LogFatalf uses fmt.Sprintf to log a templated message.
func (c *Component) LogFatalf(template string, args ...interface{}) {
	c.Logger().LogFatalf(template, args...)
}

// InitComponent is the module initializing configuration of the app.
// An App can only have one of such modules.
type InitComponent struct {
	*Component

	NonHiddenFlags []string
	// Init gets called in the initialization stage of the app.
	Init InitFunc
	// The additional configs this InitComponent brings to the app.
	AdditionalConfigs []*ConfigurationSet
}

package node

import (
	"strings"
	"sync"

	"github.com/iotaledger/hive.go/v2/events"
	"github.com/iotaledger/hive.go/v2/logger"
)

const (
	Disabled = iota
	Enabled
)

type Callback = func(plugin *Plugin)

type Plugin struct {
	Node    *Node
	Name    string
	Status  int
	Events  pluginEvents
	log     *logger.Logger
	logOnce sync.Once
	deps    interface{}
	wg      *sync.WaitGroup
}

// NewPlugin creates a new plugin with the given name, default status and callbacks.
// The last specified callback is the mandatory run callback, while all other callbacks are configure callbacks.
func NewPlugin(name string, deps interface{}, status int, callbacks ...Callback) *Plugin {
	plugin := &Plugin{
		Name:   name,
		Status: status,
		deps:   deps,
		Events: pluginEvents{
			Init:      events.NewEvent(pluginAndDepCaller),
			Configure: events.NewEvent(pluginCaller),
			Run:       events.NewEvent(pluginCaller),
		},
	}

	AddPlugin(plugin)

	switch len(callbacks) {
	case 0:
		// plugin doesn't have any callbacks (i.e. plugins that execute stuff on init())
	case 1:
		plugin.Events.Run.Attach(events.NewClosure(callbacks[0]))
	case 2:
		plugin.Events.Configure.Attach(events.NewClosure(callbacks[0]))
		plugin.Events.Run.Attach(events.NewClosure(callbacks[1]))
	default:
		panic("too many callbacks in NewPlugin(...)")
	}

	return plugin
}

func GetPluginIdentifier(name string) string {
	return strings.ToLower(strings.Replace(name, " ", "", -1))
}

// LogDebug uses fmt.Sprint to construct and log a message.
func (p *Plugin) LogDebug(args ...interface{}) {
	p.Logger().Debug(args...)
}

// LogDebugf uses fmt.Sprintf to log a templated message.
func (p *Plugin) LogDebugf(format string, args ...interface{}) {
	p.Logger().Debugf(format, args...)
}

// LogError uses fmt.Sprint to construct and log a message.
func (p *Plugin) LogError(args ...interface{}) {
	p.Logger().Error(args...)
}

// LogErrorf uses fmt.Sprintf to log a templated message.
func (p *Plugin) LogErrorf(format string, args ...interface{}) {
	p.Logger().Errorf(format, args...)
}

// LogFatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (p *Plugin) LogFatal(args ...interface{}) {
	p.Logger().Fatal(args...)
}

// LogFatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (p *Plugin) LogFatalf(format string, args ...interface{}) {
	p.Logger().Fatalf(format, args...)
}

// LogInfo uses fmt.Sprint to construct and log a message.
func (p *Plugin) LogInfo(args ...interface{}) {
	p.Logger().Info(args...)
}

// LogInfof uses fmt.Sprintf to log a templated message.
func (p *Plugin) LogInfof(format string, args ...interface{}) {
	p.Logger().Infof(format, args...)
}

// LogWarn uses fmt.Sprint to construct and log a message.
func (p *Plugin) LogWarn(args ...interface{}) {
	p.Logger().Warn(args...)
}

// LogWarnf uses fmt.Sprintf to log a templated message.
func (p *Plugin) LogWarnf(format string, args ...interface{}) {
	p.Logger().Warnf(format, args...)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func (p *Plugin) Panic(args ...interface{}) {
	p.Logger().Panic(args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func (p *Plugin) Panicf(template string, args ...interface{}) {
	p.Logger().Panicf(template, args...)
}

// Logger instantiates and returns a logger with the name of the plugin.
func (p *Plugin) Logger() *logger.Logger {
	p.logOnce.Do(func() {
		p.log = logger.NewLogger(p.Name)
	})

	return p.log
}

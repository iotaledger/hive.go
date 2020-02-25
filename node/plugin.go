package node

import (
	"strings"
	"sync"

	"github.com/iotaledger/hive.go/events"
)

const (
	Disabled = iota
	Enabled
)

type Callback = func(plugin *Plugin)

type Plugin struct {
	Node   *Node
	Name   string
	Status int
	Events pluginEvents
	wg     *sync.WaitGroup
}

// Creates a new plugin with the given name, default status and callbacks.
// The last specified callback is the mandatory run callback, while all other callbacks are configure callbacks.
func NewPlugin(name string, status int, callbacks ...Callback) *Plugin {
	plugin := &Plugin{
		Name:   name,
		Status: status,
		Events: pluginEvents{
			Init:      events.NewEvent(pluginCaller),
			Configure: events.NewEvent(pluginCaller),
			Run:       events.NewEvent(pluginCaller),
		},
	}

	AddPlugin(name, status)

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

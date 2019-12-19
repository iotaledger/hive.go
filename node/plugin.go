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
func NewPlugin(name string, status int, callback Callback, callbacks ...Callback) *Plugin {
	plugin := &Plugin{
		Name:   name,
		Status: status,
		Events: pluginEvents{
			Configure: events.NewEvent(pluginCaller),
			Run:       events.NewEvent(pluginCaller),
		},
	}

	AddPlugin(name, status)

	if len(callbacks) >= 1 {
		plugin.Events.Configure.Attach(events.NewClosure(callback))
		for _, callback = range callbacks[:len(callbacks)-1] {
			plugin.Events.Configure.Attach(events.NewClosure(callback))
		}

		plugin.Events.Run.Attach(events.NewClosure(callbacks[len(callbacks)-1]))
	} else {
		plugin.Events.Run.Attach(events.NewClosure(callback))
	}

	return plugin
}

func GetPluginIdentifier(name string) string {
	return strings.ToLower(strings.Replace(name, " ", "", -1))
}

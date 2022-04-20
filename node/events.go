package node

import (
	"github.com/iotaledger/hive.go/generics/event"
	"go.uber.org/dig"
)

var Events *NodeEvents

type NodeEvents struct {
	AddPlugin *event.Event[*AddEvent]
}

func newNodeEvents() (new *NodeEvents) {
	return &NodeEvents{
		AddPlugin: event.New[*AddEvent](),
	}
}

type AddEvent struct {
	Name   string
	Status int
}

func init() {
	Events = newNodeEvents()
}

type PluginEvents struct {
	Init      *event.Event[*InitEvent]
	Configure *event.Event[*ConfigureEvent]
	Run       *event.Event[*RunEvent]
}

func newPluginEvents() (new *PluginEvents) {
	return &PluginEvents{
		Init:      event.New[*InitEvent](),
		Configure: event.New[*ConfigureEvent](),
		Run:       event.New[*RunEvent](),
	}
}

type InitEvent struct {
	Plugin    *Plugin
	Container *dig.Container
}

type ConfigureEvent struct {
	Plugin *Plugin
}

type RunEvent struct {
	Plugin *Plugin
}

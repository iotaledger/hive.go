package node

import (
	"github.com/iotaledger/hive.go/v2/events"
	"go.uber.org/dig"
)

type pluginEvents struct {
	Init      *events.Event
	Configure *events.Event
	Run       *events.Event
}

func pluginCaller(handler interface{}, params ...interface{}) {
	handler.(func(*Plugin))(params[0].(*Plugin))
}

func pluginAndDepCaller(handler interface{}, params ...interface{}) {
	handler.(func(*Plugin, *dig.Container))(params[0].(*Plugin), params[1].(*dig.Container))
}

func pluginParameterCaller(handler interface{}, params ...interface{}) {
	handler.(func(string, int))(params[0].(string), params[1].(int))
}

var (
	Events = struct {
		AddPlugin *events.Event
	}{
		AddPlugin: events.NewEvent(pluginParameterCaller),
	}
)

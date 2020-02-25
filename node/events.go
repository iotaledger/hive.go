package node

import (
	"github.com/iotaledger/hive.go/events"
)

type pluginEvents struct {
	Init      *events.Event
	Configure *events.Event
	Run       *events.Event
}

func pluginCaller(handler interface{}, params ...interface{}) {
	handler.(func(*Plugin))(params[0].(*Plugin))
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

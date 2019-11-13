package parameter

import (
	"github.com/iotaledger/hive.go/events"
)

var Events = struct {
	AddPlugin *events.Event
}{
	AddPlugin: events.NewEvent(pluginParameterCaller),
}

func pluginParameterCaller(handler interface{}, params ...interface{}) {
	handler.(func(string, int))(params[0].(string), params[1].(int))
}

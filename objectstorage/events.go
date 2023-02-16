package objectstorage

import (
	"github.com/iotaledger/hive.go/runtime/event"
)

type Events struct {
	ObjectEvicted *event.Event2[[]byte, StorableObject]
}

func evictionEvent(handler interface{}, params ...interface{}) {
	handler.(func([]byte, StorableObject))(params[0].([]byte), params[1].(StorableObject))
}

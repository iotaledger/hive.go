package event

import (
	"github.com/iotaledger/hive.go/runtime/workerpool"
)

var Loop *workerpool.WorkerPool

func init() {
	Loop = workerpool.New("event.Loop")
	Loop.Start()
}

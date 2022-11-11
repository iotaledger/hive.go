package event

import (
	"github.com/iotaledger/hive.go/core/workerpool"
)

const loopQueueSize = 100000

var Loop *workerpool.UnboundedWorkerPool

func init() {
	Loop = workerpool.NewUnboundedWorkerPool()
	Loop.Start()
}

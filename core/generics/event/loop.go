package event

import (
	"github.com/iotaledger/hive.go/core/workerpool"
)

const loopQueueSize = 100000

var Loop *workerpool.NonBlockingWorkerPool

func init() {
	Loop = workerpool.NewNonBlockingWorkerPool(workerpool.QueueSize(loopQueueSize), workerpool.FlushTasksAtShutdown(true))
	Loop.Start()
}

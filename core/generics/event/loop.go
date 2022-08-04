package event

import (
	"github.com/iotaledger/hive.go/core/workerpool"
)

const loopQueueSize = 100000

var Loop *workerpool.BlockingQueuedWorkerPool

func init() {
	Loop = workerpool.NewBlockingQueuedWorkerPool(workerpool.QueueSize(loopQueueSize), workerpool.FlushTasksAtShutdown(true))
	Loop.Start()
}

package event

import (
	"github.com/iotaledger/hive.go/workerpool"
)

const loopQueueSize = 1000

var Loop *workerpool.BlockingQueuedWorkerPool

func init() {
	Loop = workerpool.NewBlockingQueuedWorkerPool(workerpool.QueueSize(loopQueueSize), workerpool.FlushTasksAtShutdown(true))
	Loop.Start()
}

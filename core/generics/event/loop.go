package event

import (
	workerpool2 "github.com/iotaledger/hive.go/core/workerpool"
)

const loopQueueSize = 100000

var Loop *workerpool2.BlockingQueuedWorkerPool

func init() {
	Loop = workerpool2.NewBlockingQueuedWorkerPool(workerpool2.QueueSize(loopQueueSize), workerpool2.FlushTasksAtShutdown(true))
	Loop.Start()
}

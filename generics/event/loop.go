package event

import (
	"sync"

	"github.com/iotaledger/hive.go/workerpool"
)

const loopQueueSize = 1000

type EventLoop struct {
	once sync.Once

	wp *workerpool.BlockingQueuedWorkerPool
}

var Loop EventLoop

func (l *EventLoop) GetWorkerPool() *workerpool.BlockingQueuedWorkerPool {
	l.once.Do(func() {
		l.wp = workerpool.NewBlockingQueuedWorkerPool(workerpool.QueueSize(loopQueueSize), workerpool.FlushTasksAtShutdown(true))
		l.wp.Start()
	})
	return l.wp
}

func (l *EventLoop) Submit(f func()) {
	l.GetWorkerPool().Submit(f)
}

func (l *EventLoop) TrySubmit(f func()) (added bool) {
	return l.GetWorkerPool().TrySubmit(f)
}

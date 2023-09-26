package timed

import (
	"sync"
	"time"

	"github.com/izuc/zipp.foundation/runtime/options"
)

// region TimedExecutor ////////////////////////////////////////////////////////////////////////////////////////////////

// Executor defines a scheduler that executes tasks in the background at a given time. It does not spawn any
// additional goroutines for each task and executes the tasks sequentially (in each worker).
type Executor struct {
	workerCount int

	optsMaxQueueSize int

	queue      *Queue
	shutdownWG sync.WaitGroup
}

// NewExecutor is the constructor for a timed Executor that creates a scheduler with a given number of workers that execute the
// scheduled tasks in parallel (whenever they become due).
func NewExecutor(workerCount int, opts ...options.Option[Executor]) (timedExecutor *Executor) {
	return options.Apply(&Executor{
		workerCount: workerCount,
	}, opts, func(t *Executor) {
		t.queue = NewQueue(WithMaxSize(t.optsMaxQueueSize))
		t.startBackgroundWorkers()
	})
}

// ExecuteAfter executes the given function after the given delay.
func (t *Executor) ExecuteAfter(f func(), delay time.Duration) *ScheduledTask {
	return t.ExecuteAt(f, time.Now().Add(delay))
}

// ExecuteAt executes the given function at the given time.
func (t *Executor) ExecuteAt(f func(), time time.Time) *ScheduledTask {
	return t.queue.Add(f, time)
}

// Size returns the amount of jobs that are currently scheduled for execution.
func (t *Executor) Size() int {
	return t.queue.Size()
}

// WorkerCount returns the amount of background workers that this executor uses.
func (t *Executor) WorkerCount() int {
	return t.workerCount
}

// Shutdown shuts down the TimedExecutor and waits until the executor has shutdown gracefully.
func (t *Executor) Shutdown(optionalShutdownFlags ...ShutdownFlag) {
	shutdownFlags := PanicOnModificationsAfterShutdown
	for _, optionalShutdownFlag := range optionalShutdownFlags {
		shutdownFlags |= optionalShutdownFlag
	}

	t.queue.Shutdown(shutdownFlags)

	if shutdownFlags.HasBits(DontWaitForShutdown) {
		return
	}

	t.shutdownWG.Wait()
}

// startBackgroundWorkers is an internal utility function that spawns the background workers that execute the queued tasks.
func (t *Executor) startBackgroundWorkers() {
	for i := 0; i < t.workerCount; i++ {
		t.shutdownWG.Add(1)
		go func() {
			for currentEntry := t.queue.Poll(true); currentEntry != nil; currentEntry = t.queue.Poll(true) {
				currentEntry.(func())()
			}

			t.shutdownWG.Done()
		}()
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ScheduledTask ////////////////////////////////////////////////////////////////////////////////////////////////

// ScheduledTask is.
type ScheduledTask = QueueElement

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Options///////////////////////////////////////////////////////////////////////////////////////////////////////

// WithMaxQueueSize is an ExecutorOption for the TimedExecutor that allows to specify a maxSize of the underlying queue.
func WithMaxQueueSize(maxSize int) options.Option[Executor] {
	return func(t *Executor) {
		t.optsMaxQueueSize = maxSize
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

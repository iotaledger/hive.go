package timed

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/izuc/zipp.foundation/ds/bitmask"
	"github.com/izuc/zipp.foundation/ds/generalheap"
	"github.com/izuc/zipp.foundation/runtime/options"
	"github.com/izuc/zipp.foundation/runtime/timeutil"
)

// region TimedQueue ///////////////////////////////////////////////////////////////////////////////////////////////////

// Queue represents a queue, that holds values that will only be released at a given time. The corresponding Poll
// method waits for the element to be available before it returns its value and is therefore blocking.
type Queue struct {
	heap      generalheap.Heap[HeapKey, *QueueElement]
	heapMutex sync.RWMutex

	waitCond *sync.Cond

	maxSize int

	ctx           context.Context
	ctxCancel     context.CancelFunc
	isShutdown    bool
	shutdownFlags ShutdownFlag
	shutdownMutex sync.Mutex
}

// NewQueue is the constructor for the timed Queue.
func NewQueue(opts ...options.Option[Queue]) (queue *Queue) {
	ctx, ctxCancel := context.WithCancel(context.Background())

	return options.Apply(&Queue{
		ctx:       ctx,
		ctxCancel: ctxCancel,
	}, opts, func(t *Queue) {
		t.waitCond = sync.NewCond(&t.heapMutex)
	})
}

// Add inserts a new element into the queue that can be retrieved via Poll() at the specified time.
func (t *Queue) Add(value any, scheduledTime time.Time) (addedElement *QueueElement) {
	// sanitize parameters
	if value == nil {
		panic("<nil> must not be added to the queue")
	}

	// prevent modifications of a shutdown queue
	if t.IsShutdown() {
		if t.shutdownFlags.HasBits(PanicOnModificationsAfterShutdown) {
			panic("tried to modify a shutdown TimedQueue")
		}

		return nil
	}

	// acquire locks
	t.heapMutex.Lock()

	// add new element

	element := &generalheap.HeapElement[HeapKey, *QueueElement]{
		Key: HeapKey(scheduledTime),
	}

	element.Value = &QueueElement{
		timedQueue: t,
		Value:      value,
		rawElem:    element,
		cancel:     make(chan byte),
	}
	heap.Push(&t.heap, element)

	if t.maxSize > 0 {
		// heap is bigger than maxSize now; remove the last element (furthest in the future).
		if size := t.heap.Len(); size > t.maxSize {
			heap.Remove(&t.heap, size-1)
		}
	}

	// release locks
	t.heapMutex.Unlock()

	// signal waiting goroutine to wake up
	t.waitCond.Signal()

	return element.Value
}

// Size returns the amount of elements that are currently enqueued in this queue.
func (t *Queue) Size() int {
	t.heapMutex.RLock()
	defer t.heapMutex.RUnlock()

	return len(t.heap)
}

// Shutdown terminates the queue. It accepts an optional list of shutdown flags that allows the caller to modify the
// shutdown behavior.
func (t *Queue) Shutdown(optionalShutdownFlags ...ShutdownFlag) {
	// acquire locks
	t.shutdownMutex.Lock()

	// prevent modifications of an already shutdown queue
	if t.isShutdown {
		// automatically unlock
		defer t.shutdownMutex.Unlock()

		// panic if the corresponding flag was set
		if t.shutdownFlags.HasBits(PanicOnModificationsAfterShutdown) {
			panic("tried to shutdown and already shutdown TimedQueue")
		}

		return
	}

	// mark the queue as shutdown
	t.isShutdown = true

	// store the shutdown flags
	for _, shutdownFlag := range optionalShutdownFlags {
		t.shutdownFlags |= shutdownFlag
	}

	// release the lock
	t.shutdownMutex.Unlock()

	// cancel the context to shutdown (notify waiting threads)
	t.ctxCancel()

	t.heapMutex.Lock()
	switch queuedElementsCount := len(t.heap); queuedElementsCount {
	// if the queue is empty ...
	case 0:
		// ... stop waiting for new elements
		t.waitCond.Broadcast()

	// if the queue is not empty ...
	default:
		// ... empty it if the corresponding flag was set
		if t.shutdownFlags.HasBits(CancelPendingElements) {
			for i := 0; i < queuedElementsCount; i++ {
				heap.Pop(&t.heap)
			}
		}
	}
	t.heapMutex.Unlock()
}

// IsShutdown returns true if this queue was shutdown.
func (t *Queue) IsShutdown() bool {
	t.shutdownMutex.Lock()
	defer t.shutdownMutex.Unlock()

	return t.isShutdown
}

// REVIEWED FUNCTIONS /

// Poll returns the first value of this queue. It waits for the scheduled time before returning and is therefore
// blocking. It returns nil if the queue is empty.
func (t *Queue) Poll(waitIfEmpty bool) any {
	for {
		// acquire locks
		t.heapMutex.Lock()

		// wait for elements to be queued
		for len(t.heap) == 0 {
			if !waitIfEmpty || t.IsShutdown() {
				t.heapMutex.Unlock()

				return nil
			}

			t.waitCond.Wait()
		}

		// retrieve first element
		polledElement := heap.Pop(&t.heap).(*generalheap.HeapElement[HeapKey, *QueueElement])
		// release locks
		t.heapMutex.Unlock()

		timer := time.NewTimer(time.Until(time.Time(polledElement.Key)))

		// wait for the return value to become due
		select {
		// react if the queue was shutdown while waiting
		case <-t.ctx.Done():
			// abort if the pending elements are supposed to be canceled
			if t.shutdownFlags.HasBits(CancelPendingElements) {
				timeutil.CleanupTimer(timer)
				return nil
			}

			// immediately return the value if the pending timeouts are supposed to be ignored
			if t.shutdownFlags.HasBits(IgnorePendingTimeouts) {
				timeutil.CleanupTimer(timer)
				return polledElement.Value.Value
			}

			// wait for the return value to become due
			select {
			// abort waiting for this element and return the next one instead if it was canceled
			case <-polledElement.Value.cancel:
				timeutil.CleanupTimer(timer)
				continue

			// return the result after the time is reached
			case <-timer.C:
				return polledElement.Value.Value
			}

		// abort waiting for this element and return the next one instead if it was canceled
		case <-polledElement.Value.cancel:
			timeutil.CleanupTimer(timer)
			continue

		// return the result after the time is reached
		case <-timer.C:
			return polledElement.Value.Value
		}
	}
}

// removeElement is an internal utility function that removes the given element from the queue.
func (t *Queue) removeElement(element *QueueElement) {
	// abort if the element was removed already
	if element.rawElem.Index() == -1 {
		return
	}

	// remove the element
	heap.Remove(&t.heap, element.rawElem.Index())
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region QueueElement /////////////////////////////////////////////////////////////////////////////////////////////////

// QueueElement is an element in the TimedQueue. It.
type QueueElement struct {
	// Value represents the value of the queued element.
	Value any

	timedQueue *Queue
	cancel     chan byte
	rawElem    *generalheap.HeapElement[HeapKey, *QueueElement]
}

// Cancel removed the given element from the queue and cancels its execution.
func (timedQueueElement *QueueElement) Cancel() {
	// acquire locks
	timedQueueElement.timedQueue.heapMutex.Lock()
	defer timedQueueElement.timedQueue.heapMutex.Unlock()

	// remove element from queue
	timedQueueElement.timedQueue.removeElement(timedQueueElement)

	select {
	case <-timedQueueElement.cancel:
		// channel is already closed
	default:
		// close the cancel channel to notify subscribers
		close(timedQueueElement.cancel)
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ShutdownFlags ////////////////////////////////////////////////////////////////////////////////////////////////

// ShutdownFlag defines the type of the optional shutdown flags.
type ShutdownFlag = bitmask.BitMask

const (
	// CancelPendingElements defines a shutdown flag, that causes the queue to be emptied on shutdown.
	CancelPendingElements ShutdownFlag = 1 << iota

	// IgnorePendingTimeouts defines a shutdown flag, that makes the queue ignore the timeouts of the remaining queued
	// elements. Consecutive calls to Poll will immediately return these elements.
	IgnorePendingTimeouts

	// PanicOnModificationsAfterShutdown makes the queue panic instead of ignoring consecutive writes or modifications.
	PanicOnModificationsAfterShutdown

	// DontWaitForShutdown causes the TimedExecutor to not wait for all tasks to be executed before returning from the
	// Shutdown method.
	DontWaitForShutdown ShutdownFlag = 1 << 7
)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Options///////////////////////////////////////////////////////////////////////////////////////////////////////

// WithMaxSize is an Option for the timed.Queue that allows to specify a maxSize of the queue.
func WithMaxSize(maxSize int) options.Option[Queue] {
	return func(queue *Queue) {
		queue.maxSize = maxSize
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

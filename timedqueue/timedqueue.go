package timedqueue

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/bitmask"
)

// region TimedQueue ///////////////////////////////////////////////////////////////////////////////////////////////////

// TimedQueue represents a queue, that holds values that will only be released at a given time. The corresponding Poll
// method waits for the element to be available before it returns its value and is therefore blocking.
type TimedQueue struct {
	heap      timedHeap
	heapMutex sync.RWMutex

	waitCond *sync.Cond

	maxSize int

	ctx           context.Context
	ctxCancel     context.CancelFunc
	isShutdown    bool
	shutdownFlags ShutdownFlag
	shutdownMutex sync.Mutex
}

// New is the constructor for the TimedQueue.
func New(opts ...Option) (queue *TimedQueue) {
	ctx, ctxCancel := context.WithCancel(context.Background())

	queue = &TimedQueue{
		ctx:       ctx,
		ctxCancel: ctxCancel,
	}
	queue.waitCond = sync.NewCond(&queue.heapMutex)

	for _, opt := range opts {
		opt(queue)
	}

	return
}

// Add inserts a new element into the queue that can be retrieved via Poll() at the specified time.
func (t *TimedQueue) Add(value interface{}, scheduledTime time.Time) (addedElement *QueueElement) {
	// sanitize parameters
	if value == nil {
		panic("<nil> must not be added to the queue")
	}

	// prevent modifications of a shutdown queue
	if t.IsShutdown() {
		if t.shutdownFlags.HasBits(PanicOnModificationsAfterShutdown) {
			panic("tried to modify a shutdown TimedQueue")
		}

		return
	}

	// acquire locks
	t.heapMutex.Lock()

	// add new element
	addedElement = &QueueElement{
		timedQueue:    t,
		Value:         value,
		ScheduledTime: scheduledTime,
		cancel:        make(chan byte),
		index:         0,
	}
	heap.Push(&t.heap, addedElement)

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

	return
}

// Size returns the amount of elements that are currently enqueued in this queue.
func (t *TimedQueue) Size() int {
	t.heapMutex.RLock()
	defer t.heapMutex.RUnlock()

	return len(t.heap)
}

// Shutdown terminates the queue. It accepts an optional list of shutdown flags that allows the caller to modify the
// shutdown behavior.
func (t *TimedQueue) Shutdown(optionalShutdownFlags ...ShutdownFlag) {
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
				heap.Remove(&t.heap, 0)
			}
		}
	}
	t.heapMutex.Unlock()
}

// IsShutdown returns true if this queue was shutdown.
func (t *TimedQueue) IsShutdown() bool {
	t.shutdownMutex.Lock()
	defer t.shutdownMutex.Unlock()

	return t.isShutdown
}

// REVIEWED FUNCTIONS /

// Poll returns the first value of this queue. It waits for the scheduled time before returning and is therefore
// blocking. It returns nil if the queue is empty.
func (t *TimedQueue) Poll(waitIfEmpty bool) interface{} {
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
		polledElement := heap.Remove(&t.heap, 0).(*QueueElement)

		// release locks
		t.heapMutex.Unlock()

		// wait for the return value to become due
		select {
		// react if the queue was shutdown while waiting
		case <-t.ctx.Done():
			// abort if the pending elements are supposed to be canceled
			if t.shutdownFlags.HasBits(CancelPendingElements) {
				return nil
			}

			// immediately return the value if the pending timeouts are supposed to be ignored
			if t.shutdownFlags.HasBits(IgnorePendingTimeouts) {
				return polledElement.Value
			}

			// wait for the return value to become due
			select {
			// abort waiting for this element and return the next one instead if it was canceled
			case <-polledElement.cancel:
				continue

			// return the result after the time is reached
			case <-time.After(time.Until(polledElement.ScheduledTime)):
				return polledElement.Value
			}

		// abort waiting for this element and return the next one instead if it was canceled
		case <-polledElement.cancel:
			continue

		// return the result after the time is reached
		case <-time.After(time.Until(polledElement.ScheduledTime)):
			return polledElement.Value
		}
	}
}

// removeElement is an internal utility function that removes the given element from the queue.
func (t *TimedQueue) removeElement(element *QueueElement) {
	// abort if the element was removed already
	if element.index == -1 {
		return
	}

	// remove the element
	heap.Remove(&t.heap, element.index)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region QueueElement /////////////////////////////////////////////////////////////////////////////////////////////////

// QueueElement is an element in the TimedQueue. It
type QueueElement struct {
	// Value represents the value of the queued element.
	Value interface{}

	// ScheduledTime represents the time at which the element is due.
	ScheduledTime time.Time

	timedQueue *TimedQueue
	cancel     chan byte
	index      int
}

// Cancel removed the given element from the queue and cancels its execution.
func (timedQueueElement *QueueElement) Cancel() {
	// acquire locks
	timedQueueElement.timedQueue.heapMutex.Lock()
	defer timedQueueElement.timedQueue.heapMutex.Unlock()

	// remove element from queue
	timedQueueElement.timedQueue.removeElement(timedQueueElement)

	// close the cancel channel to notify subscribers
	close(timedQueueElement.cancel)
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
)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region timedHeap ////////////////////////////////////////////////////////////////////////////////////////////////////

// timedHeap defines a heap based on times.
type timedHeap []*QueueElement

// Len is the number of elements in the collection.
func (h timedHeap) Len() int {
	return len(h)
}

// Less reports whether the element with index i should sort before the element with index j.
func (h timedHeap) Less(i, j int) bool {
	return h[i].ScheduledTime.Before(h[j].ScheduledTime)
}

// Swap swaps the elements with indexes i and j.
func (h timedHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index, h[j].index = i, j
}

// Push adds x as the last element to the heap.
func (h *timedHeap) Push(x interface{}) {
	data := x.(*QueueElement)
	*h = append(*h, data)
	data.index = len(*h) - 1
}

// Pop removes and returns the last element of the heap.
func (h *timedHeap) Pop() interface{} {
	n := len(*h)
	data := (*h)[n-1]
	(*h)[n-1] = nil // avoid memory leak
	*h = (*h)[:n-1]
	data.index = -1
	return data
}

// interface contract (allow the compiler to check if the implementation has all of the required methods).
var _ heap.Interface = &timedHeap{}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// Option //////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Option is the type for functional options of the TimedQueue.
type Option func(queue *TimedQueue)

// WithMaxSize is an Option for the TimedQueue that allows to specify a maxSize of the queue.
func WithMaxSize(maxSize int) Option {
	return func(queue *TimedQueue) {
		queue.maxSize = maxSize
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

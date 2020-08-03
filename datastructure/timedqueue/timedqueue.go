package datastructure

import (
	"container/heap"
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

	waitForNewElements sync.WaitGroup

	shutdownSignal chan byte
	isShutdown     bool
	shutdownFlags  ShutdownFlag
	shutdownMutex  sync.Mutex
}

// New is the constructor for the TimedQueue.
func New() (queue *TimedQueue) {
	queue = &TimedQueue{
		shutdownSignal: make(chan byte),
	}
	queue.waitForNewElements.Add(1)

	return
}

// Add inserts a new element into the queue that can be retrieved via Poll() at the specified time.
func (t *TimedQueue) Add(value interface{}, scheduledTime time.Time) (addedElement *TimedQueueElement) {
	// sanitize parameters
	if value == nil {
		panic("<nil> must not be added to the queue")
	}

	// prevent modifications of a shutdown queue
	if t.IsShutdown() {
		if t.shutdownFlags.HasFlag(panicOnModificationsAfterShutdown) {
			panic("tried to modify a shutdown TimedQueue")
		}

		return
	}

	// acquire locks
	t.heapMutex.Lock()
	defer t.heapMutex.Unlock()

	// mark queue as non-empty
	if len(t.heap) == 0 {
		t.waitForNewElements.Done()
	}

	// add new element
	addedElement = &TimedQueueElement{
		timedQueue: t,
		value:      value,
		time:       scheduledTime,
		cancel:     make(chan byte),
		index:      0,
	}
	heap.Push(&t.heap, addedElement)

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
		if t.shutdownFlags.HasFlag(panicOnModificationsAfterShutdown) {
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

	// close the shutdown channel (notify waiting threads)
	close(t.shutdownSignal)

	t.heapMutex.Lock()
	switch queuedElementsCount := len(t.heap); queuedElementsCount {
	// if the queue is empty ...
	case 0:
		// ... stop waiting for new elements
		t.waitForNewElements.Done()

	// if the queue is not empty ...
	default:
		// ... empty it if the corresponding flag was set
		if t.shutdownFlags.HasFlag(cancelPendingElements) {
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
	// optionally wait for new elements before continuing
	if waitIfEmpty {
		t.waitForNewElements.Wait()
	}

	// acquire locks
	t.heapMutex.Lock()

	// if the queue is empty after waiting ...
	if len(t.heap) == 0 {
		t.heapMutex.Unlock()

		// ... wait again (if the queue was not shutdown, yet and we wanted to wait)
		//
		// Note: This can happen, if multiple goroutines are simultaneously polling elements from the queue.
		//       They all wait for a new element to arrive, then one retrieves the new elements and the other goroutines
		//       will still see an empty tangle even if they waited.
		if !t.IsShutdown() && waitIfEmpty {
			return t.Poll(waitIfEmpty)
		}

		// ... abort
		return nil
	}

	// retrieve first element
	polledElement := heap.Remove(&t.heap, 0).(*TimedQueueElement)

	// update waiting for new elements wait group if necessary
	t.markEmptyQueueAsWaitingForElements()

	// release locks
	t.heapMutex.Unlock()

	// wait for the return value to become due
	select {
	// react if the queue was shutdown while waiting
	case <-t.shutdownSignal:
		// abort if the pending elements are supposed to be canceled
		if t.shutdownFlags.HasFlag(cancelPendingElements) {
			return nil
		}

		// immediately return the value if the pending timeouts are supposed to be ignored
		if t.shutdownFlags.HasFlag(ignorePendingTimeouts) {
			return polledElement.value
		}

		// wait for the return value to become due
		select {
		// abort waiting for this element and return the next one instead if it was canceled
		case <-polledElement.cancel:
			return t.Poll(waitIfEmpty)

		// return the result after the time is reached
		case <-time.After(time.Until(polledElement.time)):
			return polledElement.value
		}

	// abort waiting for this element and return the next one instead if it was canceled
	case <-polledElement.cancel:
		return t.Poll(waitIfEmpty)

	// return the result after the time is reached
	case <-time.After(time.Until(polledElement.time)):
		return polledElement.value
	}
}

// removeElement is an internal utility function that removes the given element from the queue.
func (t *TimedQueue) removeElement(element *TimedQueueElement) {
	// abort if the element was removed already
	if element.index == -1 {
		return
	}

	// remove the element
	heap.Remove(&t.heap, element.index)

	// update waiting for new elements wait group if necessary
	t.markEmptyQueueAsWaitingForElements()
}

// markEmptyQueueAsWaitingForElements is an internal utility function that marks the queue as waiting for new elements
// if it was not shutdown, yet.
func (t *TimedQueue) markEmptyQueueAsWaitingForElements() {
	if len(t.heap) == 0 {
		t.shutdownMutex.Lock()
		if !t.isShutdown {
			t.waitForNewElements.Add(1)
		}
		t.shutdownMutex.Unlock()
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region TimedQueueElement ////////////////////////////////////////////////////////////////////////////////////////////

// TimedQueueElement is an element in the TimedQueue. It
type TimedQueueElement struct {
	timedQueue *TimedQueue
	value      interface{}
	cancel     chan byte
	time       time.Time
	index      int
}

// Cancel removed the given element from the queue and cancels its execution.
func (timedQueueElement *TimedQueueElement) Cancel() {
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

// define the optional shutdown flags
const (
	// CancelPendingElements defines a shutdown flag, that causes the queue to be emptied on shutdown.
	CancelPendingElements ShutdownFlag = 1 << cancelPendingElements

	// IgnorePendingTimeouts defines a shutdown flag, that makes the queue ignore the timeouts of the remaining queued
	// elements. Consecutive calls to Poll will immediately return these elements.
	IgnorePendingTimeouts ShutdownFlag = 1 << ignorePendingTimeouts

	// PanicOnModificationsAfterShutdown makes the queue panic instead of ignoring consecutive writes or modifications.
	PanicOnModificationsAfterShutdown ShutdownFlag = 1 << panicOnModificationsAfterShutdown

	// define the bit offsets for the corresponding shutdown flags
	cancelPendingElements = iota
	ignorePendingTimeouts
	panicOnModificationsAfterShutdown
)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region timedHeap ////////////////////////////////////////////////////////////////////////////////////////////////////

// timedHeap defines a heap based on times.
type timedHeap []*TimedQueueElement

// Len is the number of elements in the collection.
func (h timedHeap) Len() int {
	return len(h)
}

// Less reports whether the element with index i should sort before the element with index j.
func (h timedHeap) Less(i, j int) bool {
	return h[i].time.Before(h[j].time)
}

// Swap swaps the elements with indexes i and j.
func (h timedHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index, h[j].index = i, j
}

// Push adds x as the last element to the heap.
func (h *timedHeap) Push(x interface{}) {
	data := x.(*TimedQueueElement)
	*h = append(*h, data)
	data.index = len(*h) - 1
}

// Pop removes and returns the last element of the heap.
func (h *timedHeap) Pop() interface{} {
	n := len(*h)
	data := (*h)[n-1]
	*h = (*h)[:n-1]
	data.index = -1
	return data
}

// interface contract (allow the compiler to check if the implementation has all of the required methods).
var _ heap.Interface = &timedHeap{}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

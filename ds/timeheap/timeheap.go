package timeheap

import (
	"container/heap"
	"sync"
	"time"
)

// timeHeapEntry is an element for TimeHeap.
type timeHeapEntry struct {
	timestamp time.Time
	count     uint64
}

// TimeHeap implements a heap sorted by time, where older elements are popped during AveragePerSecond call.
type TimeHeap struct {
	lock  *sync.Mutex
	heap  timeHeap
	total uint64
}

// NewTimeHeap creates a new TimeHeap object.
func NewTimeHeap() *TimeHeap {
	h := &TimeHeap{lock: &sync.Mutex{}}
	heap.Init(&h.heap)

	return h
}

// Add a new entry to the container with a count for the average calculation.
func (h *TimeHeap) Add(count uint64) {
	h.lock.Lock()
	defer h.lock.Unlock()
	heap.Push(&h.heap, &timeHeapEntry{timestamp: time.Now(), count: count})
	h.total += count
}

// Clear removes all elements from the container.
func (h *TimeHeap) Clear() {
	h.lock.Lock()
	defer h.lock.Unlock()

	for h.heap.Len() > 0 {
		_ = h.heap.Pop()
	}
}

// AveragePerSecond calculates the average per second of all entries in the given duration.
// older elements are removed from the container.
func (h *TimeHeap) AveragePerSecond(timeBefore time.Duration) float32 {
	h.lock.Lock()
	defer h.lock.Unlock()

	lenHeap := h.heap.Len()
	if lenHeap > 0 {
		for range lenHeap {
			//nolint:forcetypeassert // false positive, we know that the element is of type *timeHeapEntry
			oldest := heap.Pop(&h.heap).(*timeHeapEntry)

			if time.Since(oldest.timestamp) < timeBefore {
				heap.Push(&h.heap, oldest)

				break
			}

			h.total -= oldest.count
		}
	}

	return float32(h.total) / float32(timeBefore.Seconds())
}

// timedHeap defines a heap based on timeHeapEntries.
type timeHeap []*timeHeapEntry

// Len is the number of elements in the collection.
func (h timeHeap) Len() int {
	return len(h)
}

// Less reports whether the element with index i should sort before the element with index j.
func (h timeHeap) Less(i, j int) bool {
	return h[i].timestamp.Before(h[j].timestamp)
}

// Swap swaps the elements with indexes i and j.
func (h timeHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Push adds x as the last element to the heap.
func (h *timeHeap) Push(x interface{}) {
	//nolint:forcetypeassert // false positive, we know that the element is of type *timeHeapEntry
	*h = append(*h, x.(*timeHeapEntry))
}

// Pop removes and returns the last element of the heap.
func (h *timeHeap) Pop() interface{} {
	n := len(*h)
	data := (*h)[n-1]
	(*h)[n-1] = nil // avoid memory leak
	*h = (*h)[:n-1]

	return data
}

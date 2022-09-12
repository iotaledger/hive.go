package timed

import (
	"container/heap"
	"time"
)

// region Heap ////////////////////////////////////////////////////////////////////////////////////////////////////

// Heap defines a heap based on times.
type Heap[T any] []*HeapElement[T]

// Len is the number of elements in the collection.
func (h Heap[T]) Len() int {
	return len(h)
}

// Less reports whether the element with index i should sort before the element with index j.
func (h Heap[T]) Less(i, j int) bool {
	return h[i].Time.Before(h[j].Time)
}

// Swap swaps the elements with indexes i and j.
func (h Heap[T]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index, h[j].index = i, j
}

// Push adds x as the last element to the heap.
func (h *Heap[T]) Push(x interface{}) {
	data := x.(*HeapElement[T])
	*h = append(*h, data)
	data.index = len(*h) - 1
}

// Pop removes and returns the last element of the heap.
func (h *Heap[T]) Pop() interface{} {
	n := len(*h)
	data := (*h)[n-1]
	(*h)[n-1] = nil // avoid memory leak
	*h = (*h)[:n-1]
	data.index = -1
	return data
}

// interface contract (allow the compiler to check if the implementation has all the required methods).
var _ heap.Interface = &Heap[int]{}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region HeapElement /////////////////////////////////////////////////////////////////////////////////////////////////

type HeapElement[T any] struct {
	// Value represents the value of the queued element.
	Value T
	// Key represents the time of the element to be used as a key.
	Time  time.Time
	index int
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

package priorityqueue

import (
	"github.com/izuc/zipp.foundation/constraints"
)

// queue defines a heap based queue.
type queue[T constraints.Comparable[T]] []*T

// Len is the number of elements in the queue.
func (h *queue[T]) Len() int {
	return len(*h)
}

// Less reports whether the element with index i should sort before the element with index j.
func (h *queue[T]) Less(i, j int) bool {
	return (*(*h)[i]).Compare(*(*h)[j]) < 0
}

// Swap swaps the elements with indexes i and j.
func (h *queue[T]) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

// Push adds x as the last element to the heap.
func (h *queue[T]) Push(x interface{}) {
	typedX := x.(T)

	*h = append(*h, &typedX)
}

// Pop removes and returns the last element of the heap.
func (h *queue[T]) Pop() (element interface{}) {
	n := len(*h)
	element = (*h)[n-1]

	(*h)[n-1] = nil
	*h = (*h)[:n-1]

	return element
}

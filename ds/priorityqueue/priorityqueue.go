package priorityqueue

import (
	"container/heap"

	"github.com/izuc/zipp.foundation/constraints"
)

// PriorityQueue is a heap based priority queue.
type PriorityQueue[T constraints.Comparable[T]] struct {
	heap queue[T]
}

// New creates a new PriorityQueue.
func New[T constraints.Comparable[T]]() (newPriorityQueue *PriorityQueue[T]) {
	return &PriorityQueue[T]{
		heap: make(queue[T], 0),
	}
}

// Push adds x to the PriorityQueue.
func (p *PriorityQueue[T]) Push(x T) {
	heap.Push(&p.heap, x)
}

// Pop removes and returns the element with the highest priority.
func (p *PriorityQueue[T]) Pop() (element T, success bool) {
	if p.IsEmpty() {
		return element, false
	}

	return *heap.Pop(&p.heap).(*T), true
}

// Peek returns the element with the highest priority without removing it.
func (p *PriorityQueue[T]) Peek() (element T, success bool) {
	if p.IsEmpty() {
		return element, false
	}

	return *p.heap[0], true
}

// Size returns the number of elements in the PriorityQueue.
func (p *PriorityQueue[T]) Size() (size int) {
	return p.heap.Len()
}

// IsEmpty returns true if the PriorityQueue is empty.
func (p *PriorityQueue[T]) IsEmpty() (isEmpty bool) {
	return p.Size() == 0
}

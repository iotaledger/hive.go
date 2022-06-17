package priorityqueue

import (
	"container/heap"

	"github.com/iotaledger/hive.go/generics/constraints"
)

// PriorityQueue is a heap based priority queue.
type PriorityQueue[T constraints.Comparable[T]] struct {
	heap queue[T]
}

// New creates a new PriorityQueue.
func New[T constraints.Comparable[T]]() *PriorityQueue[T] {
	return &PriorityQueue[T]{
		heap: make(queue[T], 0),
	}
}

// Push adds x to the PriorityQueue.
func (p *PriorityQueue[T]) Push(x T) {
	heap.Push(&p.heap, x)
}

// Pop removes and returns the element with the highest priority.
func (p *PriorityQueue[T]) Pop() T {
	return *heap.Pop(&p.heap).(*T)
}

// PopIndex removes and returns the element at the given index.
func (p *PriorityQueue[T]) PopIndex(index int) int {
	return *heap.Remove(&p.heap, index).(*int)
}

// Peek returns the element with the highest priority without removing it.
func (p *PriorityQueue[T]) Peek() T {
	return *p.heap[0]
}

// PeekIndex returns the element at the given index without removing it.
func (p *PriorityQueue[T]) PeekIndex(index int) T {
	return *p.heap[index]
}

// Size returns the number of elements in the PriorityQueue.
func (p *PriorityQueue[T]) Size() int {
	return p.heap.Len()
}

package priorityqueue

import (
	"container/heap"
	"sync"

	"github.com/iotaledger/hive.go/generics/constraints"
)

// PriorityQueue is a heap based priority queue.
type PriorityQueue[T constraints.Comparable[T]] struct {
	heap queue[T]

	sync.RWMutex
}

// New creates a new PriorityQueue.
func New[T constraints.Comparable[T]]() *PriorityQueue[T] {
	return &PriorityQueue[T]{
		heap: make(queue[T], 0),
	}
}

// Push adds x to the PriorityQueue.
func (p *PriorityQueue[T]) Push(x T) {
	p.Lock()
	defer p.Unlock()

	heap.Push(&p.heap, x)
}

// Pop removes and returns the element with the highest priority.
func (p *PriorityQueue[T]) Pop() T {
	p.Lock()
	defer p.Unlock()

	return *heap.Pop(&p.heap).(*T)
}

// Peek returns the element with the highest priority without removing it.
func (p *PriorityQueue[T]) Peek() T {
	p.RLock()
	defer p.RUnlock()

	return *p.heap[0]
}

// Size returns the number of elements in the PriorityQueue.
func (p *PriorityQueue[T]) Size() int {
	p.RLock()
	defer p.RUnlock()

	return p.heap.Len()
}

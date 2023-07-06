package priorityqueue

import (
	"container/heap"
	"sync"

	"github.com/iotaledger/hive.go/ds/generalheap"
)

// PriorityQueue is a priority queue that sorts elements by a priority value.
type PriorityQueue[Element any, Priority generalheap.Comparable[Priority]] struct {
	// heap is the underlying heap.
	heap generalheap.Heap[Priority, Element]

	// mutex is used to synchronize access to the heap.
	mutex sync.RWMutex
}

// New creates a new PriorityQueue.
func New[Element any, Priority generalheap.Comparable[Priority]]() *PriorityQueue[Element, Priority] {
	return &PriorityQueue[Element, Priority]{
		heap: make(generalheap.Heap[Priority, Element], 0),
	}
}

// Push adds an element to the queue with the given priority.
func (p *PriorityQueue[Element, Priority]) Push(element Element, priority Priority) (remove func()) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	heapElement := &generalheap.HeapElement[Priority, Element]{
		Key:   priority,
		Value: element,
	}

	heap.Push(&p.heap, heapElement)

	return func() {
		p.mutex.Lock()
		defer p.mutex.Unlock()

		if heapElement.Index() != -1 {
			heap.Remove(&p.heap, heapElement.Index())
		}
	}
}

// Peek returns the element with the highest priority without removing it.
func (p *PriorityQueue[Element, Priority]) Peek() (element Element, exists bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if exists = p.heap.Len() != 0; exists {
		element = p.heap[0].Value
	}

	return element, exists
}

// Pop removes the element with the highest priority from the queue.
func (p *PriorityQueue[Element, Priority]) Pop() (element Element, exists bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.heap.Len() != 0 {
		if heapElement, ok := heap.Pop(&p.heap).(*generalheap.HeapElement[Priority, Element]); ok {
			element, exists = heapElement.Value, true
		}
	}

	return element, exists
}

// PopUntil removes all elements with a priority lower than the given priority from the queue.
func (p *PriorityQueue[Element, Priority]) PopUntil(priority Priority) []Element {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	values := make([]Element, 0)
	for p.heap.Len() != 0 && p.heap[0].Key.CompareTo(priority) <= 0 {
		if heapElement, ok := heap.Pop(&p.heap).(*generalheap.HeapElement[Priority, Element]); ok {
			values = append(values, heapElement.Value)
		}
	}

	return values
}

// PopAll removes all elements from the queue.
func (p *PriorityQueue[Element, Priority]) PopAll() []Element {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	values := make([]Element, 0)
	for p.heap.Len() != 0 {
		if element, ok := heap.Pop(&p.heap).(*generalheap.HeapElement[Priority, Element]); ok {
			values = append(values, element.Value)
		}
	}

	return values
}

// Size returns the number of elements in the queue.
func (p *PriorityQueue[Element, Priority]) Size() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.heap.Len()
}

// IsEmpty returns true if the queue is empty.
func (p *PriorityQueue[Element, Priority]) IsEmpty() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.heap.Len() == 0
}

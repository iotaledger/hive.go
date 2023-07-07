package timed

import (
	"time"

	"github.com/iotaledger/hive.go/ds/priorityqueue"
	"github.com/iotaledger/hive.go/lo"
)

// PriorityQueue is a priority queue whose elements are sorted by time.
type PriorityQueue[ElementType any] interface {
	// Push adds an element to the queue with the given time.
	Push(element ElementType, time time.Time)

	// Peek returns the element with the highest priority without removing it.
	Peek() (element ElementType, exists bool)

	// Pop removes the element with the highest priority from the queue.
	Pop() (element ElementType, exists bool)

	// PopUntil removes elements from the top of the queue until the given time.
	PopUntil(time time.Time) []ElementType

	// PopAll removes all elements from the queue.
	PopAll() []ElementType

	// Size returns the number of elements in the queue.
	Size() int

	// IsEmpty returns true if the queue is empty.
	IsEmpty() bool
}

// NewPriorityQueue creates a new PriorityQueue that can optionally be set to ascending order (oldest element first).
func NewPriorityQueue[T any](ascending ...bool) PriorityQueue[T] {
	if lo.First(ascending) {
		return &priorityQueueAscending[T]{
			PriorityQueue: priorityqueue.New[T, timeAscending](),
		}
	}

	return &priorityQueueDescending[T]{
		PriorityQueue: priorityqueue.New[T, timeDescending](),
	}
}

// priorityQueueAscending is a priority queue that sorts elements by their time in ascending order.
type priorityQueueAscending[T any] struct {
	*priorityqueue.PriorityQueue[T, timeAscending]
}

// Push adds an element to the queue with the given time.
func (a *priorityQueueAscending[T]) Push(element T, time time.Time) {
	a.PriorityQueue.Push(element, timeAscending(time))
}

// PopUntil removes elements from the top of the queue until the given time.
func (a *priorityQueueAscending[T]) PopUntil(time time.Time) []T {
	return a.PriorityQueue.PopUntil(timeAscending(time))
}

// timeAscending is a wrapper around time.Time that implements the Comparable interface for ascending order.
type timeAscending time.Time

// CompareTo compares the time of the receiver to the time of the given other timeAscending.
func (t timeAscending) CompareTo(other timeAscending) int {
	switch {
	case time.Time(t).Before(time.Time(other)):
		return -1
	case time.Time(t).After(time.Time(other)):
		return 1
	default:
		return 0
	}
}

// priorityQueueDescending is a priority queue that sorts elements by their time in descending order.
type priorityQueueDescending[T any] struct {
	*priorityqueue.PriorityQueue[T, timeDescending]
}

// Push adds an element to the queue with the given time.
func (d *priorityQueueDescending[T]) Push(element T, time time.Time) {
	d.PriorityQueue.Push(element, timeDescending(time))
}

// PopUntil removes elements from the top of the queue until the given time.
func (d *priorityQueueDescending[T]) PopUntil(time time.Time) []T {
	return d.PriorityQueue.PopUntil(timeDescending(time))
}

// timeDescending is a wrapper around time.Time that implements the Comparable interface for descending order.
type timeDescending time.Time

// CompareTo compares the time of the receiver to the time of the given other timeDescending.
func (t timeDescending) CompareTo(other timeDescending) int {
	switch {
	case time.Time(t).Before(time.Time(other)):
		return 1
	case time.Time(t).After(time.Time(other)):
		return -1
	default:
		return 0
	}
}

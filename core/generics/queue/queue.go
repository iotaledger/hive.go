package queue

import (
	"github.com/iotaledger/hive.go/core/datastructure/queue"
)

// Queue represents a ring buffer.
type Queue[T any] struct {
	*queue.Queue
}

// New creates a new queue with the specified capacity.
func New[T any](capacity int) *Queue[T] {
	return &Queue[T]{
		Queue: queue.New(capacity),
	}
}

// Offer adds an element to the queue and returns true.
// If the queue is full, it drops it and returns false.
func (queue *Queue[T]) Offer(element T) bool {
	return queue.Queue.Offer(element)
}

// Poll returns and removes the oldest element in the queue and true if successful.
// If returns false if the queue is empty.
func (queue *Queue[T]) Poll() (element T, success bool) {
	e, success := queue.Queue.Poll()
	if success {
		element = e.(T)
	}

	return
}

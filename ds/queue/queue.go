package queue

import "sync"

// Queue represents a ring buffer.
type Queue[T any] struct {
	ringBuffer []T
	read       int
	write      int
	capacity   int
	size       int
	mutex      sync.Mutex
}

// New creates a new queue with the specified capacity.
func New[T any](capacity int) *Queue[T] {
	return &Queue[T]{
		ringBuffer: make([]T, capacity),
		capacity:   capacity,
	}
}

// Size returns the size of the queue.
func (queue *Queue[T]) Size() int {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	return queue.size
}

// Capacity returns the capacity of the queue.
func (queue *Queue[T]) Capacity() int {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	return queue.capacity
}

func (queue *Queue[T]) ForceOffer(element T) (removedElement T, wasRemoved bool) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if queue.size == queue.capacity {
		removedElement, wasRemoved = queue.poll()
	}

	queue.ringBuffer[queue.write] = element
	queue.write = (queue.write + 1) % queue.capacity
	queue.size++

	return removedElement, wasRemoved
}

// Offer adds an element to the queue and returns true.
// If the queue is full, it drops it and returns false.
func (queue *Queue[T]) Offer(element T) bool {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if queue.size == queue.capacity {
		return false
	}

	queue.ringBuffer[queue.write] = element
	queue.write = (queue.write + 1) % queue.capacity
	queue.size++

	return true
}

// Poll returns and removes the oldest element in the queue and true if successful.
// If returns false if the queue is empty.
func (queue *Queue[T]) Poll() (element T, success bool) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	return queue.poll()
}

// poll returns and removes the oldest element in the queue and true if successful.
// If returns false if the queue is empty.
func (queue *Queue[T]) poll() (element T, success bool) {
	if success = queue.size != 0; !success {
		return
	}

	element = queue.ringBuffer[queue.read]
	var emptyElement T
	queue.ringBuffer[queue.read] = emptyElement
	queue.read = (queue.read + 1) % queue.capacity
	queue.size--

	return
}

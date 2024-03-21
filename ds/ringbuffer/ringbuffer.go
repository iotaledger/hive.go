package ringbuffer

import (
	"sync"
)

// RingBuffer is a thread-safe fixed buffer of elements with FIFO semantics.
// When the buffer is full, adding a new element overwrites the oldest element.
type RingBuffer[T any] struct {
	buffer   []T
	pos      int
	capacity int
	size     int
	mutex    sync.RWMutex
}

// NewRingBuffer creates a new RingBuffer with a maximum size of capacity.
func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	return &RingBuffer[T]{
		buffer:   make([]T, capacity),
		capacity: capacity,
	}
}

// Add adds an element to the buffer, overwriting the oldest element if the buffer is full.
func (r *RingBuffer[T]) Add(element T) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.buffer[r.pos] = element
	r.pos = (r.pos + 1) % r.capacity

	if r.size < r.capacity {
		r.size = r.size + 1
	}

	return true
}

// ToSlice returns all the elements currently in the buffer, from newest to oldest.
func (r *RingBuffer[T]) ToSlice() []T {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := make([]T, r.size)
	i := r.pos - 1
	if i < 0 {
		i = r.capacity - 1
	}
	for j := range r.size {
		result[j] = r.buffer[i]
		i--
		if i < 0 {
			i = r.capacity - 1
		}
	}

	return result
}

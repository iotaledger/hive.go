package set

import (
	"sync"

	"github.com/izuc/zipp.foundation/ds/types"
)

// threadSafeSet implements a thread safe Set.
type threadSafeSet[T comparable] struct {
	set   *simpleSet[T]
	mutex sync.RWMutex
}

// newThreadSafeSet returns a new thread safe Set.
func newThreadSafeSet[T comparable]() *threadSafeSet[T] {
	s := new(threadSafeSet[T])
	s.initialize()
	return s
}

func (s *threadSafeSet[T]) initialize() {
	s.set = newSimpleSet[T]()
}

// Add adds a new element to the Set and returns true if the element was not present in the set before.
func (s *threadSafeSet[T]) Add(element T) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.set.Add(element)
}

// Delete removes the element from the Set and returns true if it did exist.
func (s *threadSafeSet[T]) Delete(element T) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.set.Delete(element)
}

// Has returns true if the element exists in the Set.
func (s *threadSafeSet[T]) Has(element T) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.set.Has(element)
}

// ForEach iterates through the set and calls the callback for every element.
func (s *threadSafeSet[T]) ForEach(callback func(element T)) {
	s.mutex.RLock()
	copiedElements := make(map[T]types.Empty)
	s.set.ForEach(func(element T) {
		copiedElements[element] = types.Void
	})
	s.mutex.RUnlock()

	for element := range copiedElements {
		callback(element)
	}
}

// Clear removes all elements from the Set.
func (s *threadSafeSet[T]) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.set = newSimpleSet[T]()
}

// Size returns the size of the Set.
func (s *threadSafeSet[T]) Size() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.set.Size()
}

func (s *threadSafeSet[T]) Encode() ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.set.Encode()
}

func (s *threadSafeSet[T]) Decode(b []byte) (bytesRead int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.initialize()

	return s.set.Decode(b)
}

// code contract - make sure the type implements the interface.
var _ Set[int] = &threadSafeSet[int]{}

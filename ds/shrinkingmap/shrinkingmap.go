package shrinkingmap

import "sync"

// the default options applied to the ShrinkingMap.
var defaultOptions = []Option{
	WithShrinkingThresholdRatio(10.0),
	WithShrinkingThresholdCount(100),
}

// Options define options for a ShrinkingMap.
type Options struct {
	// The ratio between the amount of deleted keys and
	// the current map's size before shrinking is triggered.
	shrinkingThresholdRatio float32
	// The count of deletions that triggers shrinking of the map.
	shrinkingThresholdCount int
}

// applies the given Option.
func (so *Options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(so)
	}
}

// WithShrinkingThresholdRatio defines the ratio between the amount
// of deleted keys and the current map's size before shrinking is triggered.
func WithShrinkingThresholdRatio(ratio float32) Option {
	return func(opts *Options) {
		opts.shrinkingThresholdRatio = ratio
	}
}

// WithShrinkingThresholdCount defines the count of
// deletions that triggers shrinking of the map.
func WithShrinkingThresholdCount(count int) Option {
	return func(opts *Options) {
		opts.shrinkingThresholdCount = count
	}
}

// Option is a function setting an Options option.
type Option func(opts *Options)

// ShrinkingMap provides a non concurrent-safe map
// that shrinks if certain conditions are met (AND condition).
// Default values are:
// - ShrinkingThresholdRatio: 10.0	(set to 0.0 to disable)
// - ShrinkingThresholdCount: 100	(set to 0 to disable).
type ShrinkingMap[K comparable, V any] struct {
	m           map[K]V
	deletedKeys int

	// holds the map options.
	opts *Options

	mutex sync.RWMutex
}

// New returns a new ShrinkingMap.
func New[K comparable, V any](opts ...Option) *ShrinkingMap[K, V] {
	mapOpts := &Options{}
	mapOpts.apply(defaultOptions...)
	mapOpts.apply(opts...)

	shrinkingMap := &ShrinkingMap[K, V]{
		m:    make(map[K]V),
		opts: mapOpts,
	}

	return shrinkingMap
}

// Set adds a key-value pair to the map. It returns true if the key was created.
func (s *ShrinkingMap[K, V]) Set(key K, value V) (wasCreated bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.m[key]
	s.m[key] = value

	return !exists
}

// Get returns the value mapped to the given key, and the boolean flag that indicated if the key exists.
func (s *ShrinkingMap[K, V]) Get(key K) (value V, exists bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, exists = s.m[key]

	return
}

// GetOrCreate returns the value mapped to the given key and the boolean flag that indicated if the values was created.
// If the value does not exist the passed func will be called and the provided value will be set.
func (s *ShrinkingMap[K, V]) GetOrCreate(key K, defaultValueFunc func() V) (value V, created bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if existingValue, exists := s.m[key]; exists {
		return existingValue, false
	}

	value = defaultValueFunc()
	s.m[key] = value

	return value, true
}

// Has returns if an entry with the given key exists.
func (s *ShrinkingMap[K, V]) Has(key K) (has bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, has = s.m[key]

	return
}

// ForEachKey iterates through the map and calls the consumer for every element.
// Returning false from this function indicates to abort the iteration.
func (s *ShrinkingMap[K, V]) ForEachKey(callback func(K) bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for k := range s.m {
		if !callback(k) {
			return
		}
	}
}

// ForEach iterates through the map and calls the consumer for every element.
// Returning false from this function indicates to abort the iteration.
func (s *ShrinkingMap[K, V]) ForEach(callback func(K, V) bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for k, v := range s.m {
		if !callback(k, v) {
			return
		}
	}
}

// Pop removes the first element from the map and returns it.
func (s *ShrinkingMap[K, V]) Pop() (key K, value V, exists bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for k, v := range s.m {
		s.delete(k)

		return k, v, true
	}

	return
}

// Size returns the number of entries in the map.
func (s *ShrinkingMap[K, V]) Size() (size int) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.m)
}

// IsEmpty returns if the map is empty.
func (s *ShrinkingMap[K, V]) IsEmpty() (empty bool) {
	return s.Size() == 0
}

// Delete removes the entry with the given key, and possibly
// shrinks the map if the shrinking conditions have been reached.
func (s *ShrinkingMap[K, V]) Delete(key K) (deleted bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.delete(key)
}

// delete removes the entry with the given key, and possibly
// shrinks the map if the shrinking conditions have been reached.
// This does not lock the mutex.
func (s *ShrinkingMap[K, V]) delete(key K) (deleted bool) {
	if _, deleted = s.m[key]; !deleted {
		return false
	}

	s.deletedKeys++
	delete(s.m, key)

	if s.shouldShrink() {
		s.shrink()
	}

	return true
}

// AsMap returns the shrinking map as a regular map.
func (s *ShrinkingMap[K, V]) AsMap() (asMap map[K]V) {
	asMap = make(map[K]V)
	s.ForEach(func(k K, v V) bool {
		asMap[k] = v
		return true
	})
	return
}

// shouldShrink checks if the conditions to shrink the map are met.
func (s *ShrinkingMap[K, V]) shouldShrink() bool {
	size := len(s.m)

	// check if one of the conditions was defined, otherwise never shrink
	if !(s.opts.shrinkingThresholdRatio != 0.0 || s.opts.shrinkingThresholdCount != 0) {
		return false
	}

	if s.opts.shrinkingThresholdRatio != 0.0 {
		// ratio was defined

		// check for division by zero
		if size == 0 {
			return false
		}

		if float32(s.deletedKeys)/float32(size) < s.opts.shrinkingThresholdRatio {
			// condition not reached
			return false
		}
	}

	if s.opts.shrinkingThresholdCount != 0 {
		// count was defined

		if s.deletedKeys < s.opts.shrinkingThresholdCount {
			// condition not reached
			return false
		}
	}

	return true
}

// Shrink shrinks the map.
func (s *ShrinkingMap[K, V]) Shrink() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.shrink()
}

// shrink shrinks the map without locking the mutex.
func (s *ShrinkingMap[K, V]) shrink() {
	newMap := make(map[K]V, len(s.m))

	for k, v := range s.m {
		newMap[k] = v
	}

	s.deletedKeys = 0
	s.m = newMap
}

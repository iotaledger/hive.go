package shrinkingmap

// ShrinkingMap provides a non concurrent-safe map that shrinks if the amount of deleted keys is larger than the current
// size than a configurable threshold.
type ShrinkingMap[K comparable, V any] struct {
	m                  map[K]V
	shrinkingThreshold int
	deletedKeys        int
}

// New returns a new ShrinkingMap.
func New[K comparable, V any](shrinkingThreshold ...int) (new *ShrinkingMap[K, V]) {
	new = &ShrinkingMap[K, V]{
		m:                  make(map[K]V),
		shrinkingThreshold: 10,
	}

	if len(shrinkingThreshold) > 0 {
		new.shrinkingThreshold = shrinkingThreshold[0]
	}

	return new
}

// Set adds a key-value pair to the map. It returns true if the key existed.
func (s *ShrinkingMap[K, V]) Set(key K, value V) (updated bool) {
	_, updated = s.m[key]
	s.m[key] = value

	return updated
}

// Get returns the value mapped to the given key, and the boolean flag that indicated is the key exists.
func (s *ShrinkingMap[K, V]) Get(key K) (value V, exists bool) {
	value, exists = s.m[key]
	return
}

// Has returns if an entry with the given key exists.
func (s *ShrinkingMap[K, V]) Has(key K) (has bool) {
	_, has = s.m[key]
	return
}

// Size returns the number of entries in the map.
func (s *ShrinkingMap[K, V]) Size() (size int) {
	return len(s.m)
}

// IsEmpty returns if the map is empty.
func (s *ShrinkingMap[K, V]) IsEmpty() (empty bool) {
	return s.Size() == 0
}

// Delete removes the entry with the given key, and possibly shrinks the map if the shrinking threshold has been reached.
func (s *ShrinkingMap[K, V]) Delete(key K) (deleted bool) {
	if _, deleted = s.m[key]; !deleted {
		return false
	}

	s.deletedKeys++
	delete(s.m, key)

	if s.deletedKeys/len(s.m) >= s.shrinkingThreshold {
		s.Shrink()
	}

	return true
}

// Shrink shrinks the map.
func (s *ShrinkingMap[K, V]) Shrink() {
	newMap := make(map[K]V)

	for k, v := range s.m {
		newMap[k] = v
	}

	s.deletedKeys = 0
	s.m = newMap
}

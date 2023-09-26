package randommap

import (
	"math/rand"
	"sync"

	"github.com/izuc/zipp.foundation/ds/shrinkingmap"
)

type randomMapEntry[K comparable, V any] struct {
	key      K
	value    V
	keyIndex int
}

// RandomMap defines a map with extended ability to return a random entry.
type RandomMap[K comparable, V any] struct {
	rawMap *shrinkingmap.ShrinkingMap[K, *randomMapEntry[K, V]]
	keys   []K
	mutex  sync.RWMutex
}

// New creates a new random map.
func New[K comparable, V any](opts ...shrinkingmap.Option) *RandomMap[K, V] {
	return &RandomMap[K, V]{
		rawMap: shrinkingmap.New[K, *randomMapEntry[K, V]](opts...),
		keys:   make([]K, 0),
	}
}

// Set associates the specified value with the specified key.
// If the association already exists, it updates the value.
func (rmap *RandomMap[K, V]) Set(key K, value V) {
	rmap.mutex.Lock()
	defer rmap.mutex.Unlock()

	if entry, exists := rmap.rawMap.Get(key); exists {
		entry.value = value
	} else {
		rmap.rawMap.Set(key, &randomMapEntry[K, V]{
			key:      key,
			value:    value,
			keyIndex: rmap.rawMap.Size(),
		})

		rmap.keys = append(rmap.keys, key)
	}
}

// Get returns the value to which the specified key is mapped.
func (rmap *RandomMap[K, V]) Get(key K) (result V, exists bool) {
	rmap.mutex.RLock()
	defer rmap.mutex.RUnlock()

	if entry, entryExists := rmap.rawMap.Get(key); entryExists {
		result = entry.value
		exists = entryExists
	}

	return
}

// Has returns a boolean value indicating whether it exists in the map.
func (rmap *RandomMap[K, V]) Has(key K) bool {
	rmap.mutex.RLock()
	defer rmap.mutex.RUnlock()

	return rmap.rawMap.Has(key)
}

// Delete removes the mapping for the specified key in the map.
func (rmap *RandomMap[K, V]) Delete(key K) (value V, deleted bool) {
	rmap.mutex.Lock()
	defer rmap.mutex.Unlock()

	if entry, exists := rmap.rawMap.Get(key); exists {
		if entry.keyIndex != len(rmap.keys) {
			// move the last key to the position of the deleted key to shrink the slice
			oldKeyIndex := entry.keyIndex
			movedKeyIndex := len(rmap.keys) - 1

			movedKey := rmap.keys[movedKeyIndex]

			movedEntry, _ := rmap.rawMap.Get(movedKey)
			movedEntry.keyIndex = oldKeyIndex

			rmap.keys[oldKeyIndex] = movedKey

			var defaultKey K
			rmap.keys[movedKeyIndex] = defaultKey
		}

		rmap.keys = rmap.keys[:len(rmap.keys)-1]

		return entry.value, rmap.rawMap.Delete(key)
	}

	return
}

// Size returns the number of key-value mappings in the map.
func (rmap *RandomMap[K, V]) Size() int {
	rmap.mutex.RLock()
	defer rmap.mutex.RUnlock()

	return rmap.rawMap.Size()
}

// ForEach iterates through the elements in the map and calls the consumer function for each element.
func (rmap *RandomMap[K, V]) ForEach(consumer func(key K, value V) bool) {
	rmap.mutex.RLock()
	defer rmap.mutex.RUnlock()

	rmap.forEach(consumer)
}

// RandomKey returns a random key from the map.
func (rmap *RandomMap[K, V]) RandomKey() (defaultValue K, exists bool) {
	rmap.mutex.RLock()
	defer rmap.mutex.RUnlock()

	if len(rmap.keys) == 0 {
		return defaultValue, false
	}

	return rmap.randomKey(), true
}

// RandomEntry returns a random value from the map.
func (rmap *RandomMap[K, V]) RandomEntry() (defaultValue V, exists bool) {
	rmap.mutex.RLock()
	defer rmap.mutex.RUnlock()

	if rmap.rawMap.Size() == 0 {
		return defaultValue, false
	}

	if entry, exists := rmap.rawMap.Get(rmap.randomKey()); exists {
		return entry.value, true
	}

	return defaultValue, false
}

// RandomUniqueEntries returns n random and unique values from the map.
// When count is equal or bigger than the size of the random map, the every entry in the map is returned.
func (rmap *RandomMap[K, V]) RandomUniqueEntries(count int) (results []V) {
	rmap.mutex.RLock()
	defer rmap.mutex.RUnlock()

	// zero or negative count results in empty result
	if count < 1 {
		return results
	}

	// can only return as many as there are in the map
	if rmap.rawMap.Size() <= count {
		results = make([]V, 0, rmap.rawMap.Size())
		rmap.forEach(func(key K, value V) bool {
			results = append(results, value)
			return true
		})

		return results
	}

	// helper to keep track of already seen keys
	results = make([]V, 0, count)
	randomOrder := rand.Perm(len(rmap.keys))

	// there has to be at least (count + 1) key value pairs in the map
	for idx := 0; idx < len(randomOrder) && len(results) < count; idx++ {
		randomKey := rmap.keys[randomOrder[idx]]
		if randomEntry, exists := rmap.rawMap.Get(randomKey); exists {
			results = append(results, randomEntry.value)
		}
	}

	return results
}

// Keys returns the list of keys stored in the RandomMap.
func (rmap *RandomMap[K, V]) Keys() (result []K) {
	rmap.mutex.RLock()
	defer rmap.mutex.RUnlock()

	result = make([]K, rmap.rawMap.Size())
	copy(result, rmap.keys)

	return
}

// randomKey gets a random key from the map.
func (rmap *RandomMap[K, V]) randomKey() (result K) {
	//nolint:gosec // we do not care about weak random numbers here
	return rmap.keys[rand.Intn(rmap.rawMap.Size())]
}

// forEach executes a function for all key-value pairs in the map.
func (rmap *RandomMap[K, V]) forEach(consumer func(key K, value V) bool) {
	rmap.rawMap.ForEach(func(key K, entry *randomMapEntry[K, V]) bool {
		return consumer(key, entry.value)
	})
}

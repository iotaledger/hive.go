package randommap

import (
	"math/rand"
	"sync"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
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
func (r *RandomMap[K, V]) Set(key K, value V) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if entry, exists := r.rawMap.Get(key); exists {
		entry.value = value
	} else {
		r.rawMap.Set(key, &randomMapEntry[K, V]{
			key:      key,
			value:    value,
			keyIndex: r.rawMap.Size(),
		})

		r.keys = append(r.keys, key)
	}
}

// Get returns the value to which the specified key is mapped.
func (r *RandomMap[K, V]) Get(key K) (result V, exists bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if entry, entryExists := r.rawMap.Get(key); entryExists {
		result = entry.value
		exists = entryExists
	}

	return
}

// Has returns a boolean value indicating whether it exists in the map.
func (r *RandomMap[K, V]) Has(key K) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.rawMap.Has(key)
}

// Delete removes the mapping for the specified key in the map.
func (r *RandomMap[K, V]) Delete(key K) (value V, deleted bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if entry, exists := r.rawMap.Get(key); exists {
		if entry.keyIndex != len(r.keys) {
			// move the last key to the position of the deleted key to shrink the slice
			oldKeyIndex := entry.keyIndex
			movedKeyIndex := len(r.keys) - 1

			movedKey := r.keys[movedKeyIndex]

			movedEntry, _ := r.rawMap.Get(movedKey)
			movedEntry.keyIndex = oldKeyIndex

			r.keys[oldKeyIndex] = movedKey

			var defaultKey K
			r.keys[movedKeyIndex] = defaultKey
		}

		r.keys = r.keys[:len(r.keys)-1]

		return entry.value, r.rawMap.Delete(key)
	}

	return
}

// Size returns the number of key-value mappings in the map.
func (r *RandomMap[K, V]) Size() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.rawMap.Size()
}

// ForEach iterates through the elements in the map and calls the consumer function for each element.
func (r *RandomMap[K, V]) ForEach(consumer func(key K, value V) bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	r.forEach(consumer)
}

// RandomKey returns a random key from the map.
func (r *RandomMap[K, V]) RandomKey() (defaultValue K, exists bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if len(r.keys) == 0 {
		return defaultValue, false
	}

	return r.randomKey(), true
}

// RandomEntry returns a random value from the map.
func (r *RandomMap[K, V]) RandomEntry() (defaultValue V, exists bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if r.rawMap.Size() == 0 {
		return defaultValue, false
	}

	if entry, exists := r.rawMap.Get(r.randomKey()); exists {
		return entry.value, true
	}

	return defaultValue, false
}

// RandomUniqueEntries returns n random and unique values from the map.
// When count is equal or bigger than the size of the random map, the every entry in the map is returned.
func (r *RandomMap[K, V]) RandomUniqueEntries(count int) (results []V) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// zero or negative count results in empty result
	if count < 1 {
		return results
	}

	// can only return as many as there are in the map
	if r.rawMap.Size() <= count {
		results = make([]V, 0, r.rawMap.Size())
		//nolint:revive // better be explicit here
		r.forEach(func(key K, value V) bool {
			results = append(results, value)
			return true
		})

		return results
	}

	// helper to keep track of already seen keys
	results = make([]V, 0, count)
	randomOrder := rand.Perm(len(r.keys))

	// there has to be at least (count + 1) key value pairs in the map
	for idx := 0; idx < len(randomOrder) && len(results) < count; idx++ {
		randomKey := r.keys[randomOrder[idx]]
		if randomEntry, exists := r.rawMap.Get(randomKey); exists {
			results = append(results, randomEntry.value)
		}
	}

	return results
}

// Keys returns the list of keys stored in the RandomMap.
func (r *RandomMap[K, V]) Keys() (result []K) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result = make([]K, r.rawMap.Size())
	copy(result, r.keys)

	return
}

// Values returns the list of values stored in the RandomMap.
func (r *RandomMap[K, V]) Values() (result []V) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	//nolint:revive // better be explicit here
	r.forEach(func(key K, value V) bool {
		result = append(result, value)
		return true
	})

	return
}

// randomKey gets a random key from the map.
func (r *RandomMap[K, V]) randomKey() (result K) {
	//nolint:gosec // we do not care about weak random numbers here
	return r.keys[rand.Intn(r.rawMap.Size())]
}

// forEach executes a function for all key-value pairs in the map.
func (r *RandomMap[K, V]) forEach(consumer func(key K, value V) bool) {
	r.rawMap.ForEach(func(key K, entry *randomMapEntry[K, V]) bool {
		return consumer(key, entry.value)
	})
}

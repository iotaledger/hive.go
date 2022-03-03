package randommap

import (
	"github.com/iotaledger/hive.go/datastructure/randommap"
)

// RandomMap defines a  map with extended ability to return a random entry.
type RandomMap[K comparable, V any] struct {
	*randommap.RandomMap
}

// New creates a new random map
func New[K comparable, V any]() *RandomMap[K, V] {
	return &RandomMap[K, V]{
		RandomMap: randommap.New(),
	}
}

// Set associates the specified value with the specified key.
// If the association already exists, it updates the value.
func (rmap *RandomMap[K, V]) Set(key K, value V) (updated bool) {
	return rmap.RandomMap.Set(key, value)
}

// Get returns the value to which the specified key is mapped.
func (rmap *RandomMap[K, V]) Get(key K) (result V, exists bool) {
	value, exists := rmap.RandomMap.Get(key)
	if exists {
		result = value.(V)
	}
	return
}

// Delete removes the mapping for the specified key in the map.
func (rmap *RandomMap[K, V]) Delete(key K) (result V, exists bool) {
	value, exists := rmap.RandomMap.Delete(key)
	if exists {
		result = value.(V)
	}
	return
}

// ForEach iterates through the elements in the map and calls the consumer function for each element.
func (rmap *RandomMap[K, V]) ForEach(consumer func(key K, value V)) {
	rmap.RandomMap.ForEach(func(key interface{}, value interface{}) {
		consumer(key.(K), value.(V))
	})
}

// RandomKey returns a random key from the map.
func (rmap *RandomMap[K, V]) RandomKey() (result K, exists bool) {
	randKey := rmap.RandomMap.RandomKey()
	if randKey != nil {
		return randKey.(K), true
	}
	return result, false
}

// RandomEntry returns a random value from the map.
func (rmap *RandomMap[K, V]) RandomEntry() (result V, exists bool) {
	randEntry := rmap.RandomMap.RandomEntry()
	if randEntry != nil {
		return randEntry.(V), true
	}
	return result, false

}

// RandomUniqueEntries returns n random and unique values from the map.
// When count is equal or bigger than the size of the random map, the every entry in the map is returned.
func (rmap *RandomMap[K, V]) RandomUniqueEntries(count int) (results []V) {
	entries := rmap.RandomMap.RandomUniqueEntries(count)
	results = make([]V, len(entries))
	for i, v := range entries {
		results[i] = v.(V)
	}
	return
}

// Keys returns the list of keys stored in the RandomMap.
func (rmap *RandomMap[K, V]) Keys() (result []K) {
	entries := rmap.RandomMap.Keys()
	result = make([]K, len(entries))
	for i, v := range entries {
		result[i] = v.(K)
	}
	return
}

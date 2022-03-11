package orderedmap

import (
	"github.com/iotaledger/hive.go/datastructure/orderedmap"
)

// OrderedMap provides a concurrent-safe ordered map.
type OrderedMap[K comparable, V any] struct {
	*orderedmap.OrderedMap
}

// New returns a new *OrderedMap.
func New[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		OrderedMap: orderedmap.New(),
	}
}

// Head returns the first map entry.
func (orderedMap *OrderedMap[K, V]) Head() (key K, value V, exists bool) {
	k, v, exists := orderedMap.OrderedMap.Head()
	if exists {
		key = k.(K)
		value = v.(V)
	}
	return
}

// Tail returns the last map entry.
func (orderedMap *OrderedMap[K, V]) Tail() (key K, value V, exists bool) {
	k, v, exists := orderedMap.OrderedMap.Tail()
	if exists {
		key = k.(K)
		value = v.(V)
	}
	return
}

// Has returns if an entry with the given key exists.
func (orderedMap *OrderedMap[K, V]) Has(key K) (has bool) {
	return orderedMap.OrderedMap.Has(key)
}

// Get returns the value mapped to the given key if exists.
func (orderedMap *OrderedMap[K, V]) Get(key K) (value V, exists bool) {
	v, exists := orderedMap.OrderedMap.Get(key)
	if exists {
		value = v.(V)
	}
	return
}

// Set adds a key-value pair to the orderedMap. It returns false if the same pair already exists.
func (orderedMap *OrderedMap[K, V]) Set(key K, newValue V) bool {
	return orderedMap.OrderedMap.Set(key, newValue)
}

// ForEach iterates through the orderedMap and calls the consumer function for every element.
// The iteration can be aborted by returning false in the consumer.
func (orderedMap *OrderedMap[K, V]) ForEach(consumer func(key K, value V) bool) bool {
	return orderedMap.OrderedMap.ForEach(func(key, value interface{}) bool {
		return consumer(key.(K), value.(V))
	})
}

// ForEachReverse iterates through the orderedMap in reverse order and calls the consumer function for every element.
// The iteration can be aborted by returning false in the consumer.
func (orderedMap *OrderedMap[K, V]) ForEachReverse(consumer func(key K, value V) bool) bool {
	return orderedMap.OrderedMap.ForEachReverse(func(key, value interface{}) bool {
		return consumer(key.(K), value.(V))
	})

}

// Delete deletes the given key (and related value) from the orderedMap.
// It returns false if the key is not found.
func (orderedMap *OrderedMap[K, V]) Delete(key K) bool {
	return orderedMap.OrderedMap.Delete(key)
}

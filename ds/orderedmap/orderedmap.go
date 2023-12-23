package orderedmap

import (
	"sync"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
)

// OrderedMap provides a concurrent-safe ordered map.
type OrderedMap[K comparable, V any] struct {
	head       *Element[K, V]
	tail       *Element[K, V]
	dictionary *shrinkingmap.ShrinkingMap[K, *Element[K, V]]
	size       int
	mutex      sync.RWMutex
}

// New returns a new *OrderedMap.
func New[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		dictionary: shrinkingmap.New[K, *Element[K, V]](),
	}
}

// Head returns the first map entry.
func (o *OrderedMap[K, V]) Head() (key K, value V, exists bool) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	if exists = o.head != nil; !exists {
		return
	}
	key = o.head.key
	value = o.head.value

	return
}

// Tail returns the last map entry.
func (o *OrderedMap[K, V]) Tail() (key K, value V, exists bool) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	if exists = o.tail != nil; !exists {
		return
	}
	key = o.tail.key
	value = o.tail.value

	return
}

// Has returns if an entry with the given key exists.
func (o *OrderedMap[K, V]) Has(key K) (has bool) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	return o.dictionary.Has(key)
}

// Get returns the value mapped to the given key if exists.
func (o *OrderedMap[K, V]) Get(key K) (value V, exists bool) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	orderedMapElement, orderedMapElementExists := o.dictionary.Get(key)
	if !orderedMapElementExists {
		var result V
		return result, false
	}

	return orderedMapElement.value, true
}

// Set adds a key-value pair to the orderedMap.
func (o *OrderedMap[K, V]) Set(key K, newValue V) (previousValue V, previousValueExisted bool) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if oldValue, oldValueExists := o.dictionary.Get(key); oldValueExists {
		previousValue = oldValue.value
		oldValue.value = newValue

		return previousValue, true
	}

	newElement := new(Element[K, V])
	newElement.key = key
	newElement.value = newValue

	if o.head == nil {
		o.head = newElement
	} else {
		o.tail.next = newElement
		newElement.prev = o.tail
	}
	o.tail = newElement
	o.size++

	o.dictionary.Set(key, newElement)

	return previousValue, false
}

// ForEach iterates through the orderedMap and calls the consumer function for every element.
// The iteration can be aborted by returning false in the consumer.
func (o *OrderedMap[K, V]) ForEach(consumer func(key K, value V) bool) bool {
	if o == nil {
		return true
	}

	o.mutex.RLock()
	currentEntry := o.head
	o.mutex.RUnlock()

	for currentEntry != nil {
		if !consumer(currentEntry.key, currentEntry.value) {
			return false
		}

		o.mutex.RLock()
		currentEntry = currentEntry.next
		o.mutex.RUnlock()
	}

	return true
}

// ForEachReverse iterates through the orderedMap in reverse order and calls the consumer function for every element.
// The iteration can be aborted by returning false in the consumer.
func (o *OrderedMap[K, V]) ForEachReverse(consumer func(key K, value V) bool) bool {
	if o == nil {
		return true
	}

	o.mutex.RLock()
	currentEntry := o.tail
	o.mutex.RUnlock()

	for currentEntry != nil {
		if !consumer(currentEntry.key, currentEntry.value) {
			return false
		}

		o.mutex.RLock()
		currentEntry = currentEntry.prev
		o.mutex.RUnlock()
	}

	return true
}

// Clear removes all elements from the OrderedMap.
func (o *OrderedMap[K, V]) Clear() {
	if o == nil {
		return
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.head = nil
	o.tail = nil
	o.size = 0
	o.dictionary = shrinkingmap.New[K, *Element[K, V]]()
}

// Delete deletes the given key (and related value) from the orderedMap.
// It returns false if the key is not found.
func (o *OrderedMap[K, V]) Delete(key K) bool {
	if _, valueExists := o.Get(key); !valueExists {
		return false
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	value, valueExists := o.dictionary.Get(key)
	if !valueExists {
		return false
	}

	o.dictionary.Delete(key)
	o.size--

	if value.prev != nil {
		value.prev.next = value.next
	} else {
		o.head = value.next
	}

	if value.next != nil {
		value.next.prev = value.prev
	} else {
		o.tail = value.prev
	}

	return true
}

// Size returns the size of the orderedMap.
func (o *OrderedMap[K, V]) Size() int {
	if o == nil {
		return 0
	}

	o.mutex.RLock()
	defer o.mutex.RUnlock()

	return o.size
}

// IsEmpty returns a boolean value indicating whether the map empty.
func (o *OrderedMap[K, V]) IsEmpty() bool {
	return o.Size() == 0
}

// Clone returns a copy of the orderedMap.
func (o *OrderedMap[K, V]) Clone() *OrderedMap[K, V] {
	if o == nil {
		return nil
	}

	cloned := New[K, V]()

	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for currentEntry := o.head; currentEntry != nil; currentEntry = currentEntry.next {
		cloned.Set(currentEntry.key, currentEntry.value)
	}

	return cloned
}

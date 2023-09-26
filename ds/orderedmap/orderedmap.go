package orderedmap

import (
	"context"
	"sync"

	"golang.org/x/xerrors"

	"github.com/izuc/zipp.foundation/serializer/v2"
	"github.com/izuc/zipp.foundation/serializer/v2/serix"
)

// OrderedMap provides a concurrent-safe ordered map.
type OrderedMap[K comparable, V any] struct {
	head       *Element[K, V]
	tail       *Element[K, V]
	dictionary map[K]*Element[K, V]
	size       int
	mutex      sync.RWMutex
}

// New returns a new *OrderedMap.
func New[K comparable, V any]() *OrderedMap[K, V] {
	orderedMap := new(OrderedMap[K, V])
	orderedMap.Initialize()

	return orderedMap
}

// Initialize returns the first map entry.
func (o *OrderedMap[K, V]) Initialize() {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.dictionary = make(map[K]*Element[K, V])
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

	_, has = o.dictionary[key]

	return
}

// Get returns the value mapped to the given key if exists.
func (o *OrderedMap[K, V]) Get(key K) (value V, exists bool) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	orderedMapElement, orderedMapElementExists := o.dictionary[key]
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

	if oldValue, oldValueExists := o.dictionary[key]; oldValueExists {
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

	o.dictionary[key] = newElement

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
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.head = nil
	o.tail = nil
	o.size = 0
	o.dictionary = make(map[K]*Element[K, V])
}

// Delete deletes the given key (and related value) from the orderedMap.
// It returns false if the key is not found.
func (o *OrderedMap[K, V]) Delete(key K) bool {
	if _, valueExists := o.Get(key); !valueExists {
		return false
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	value, valueExists := o.dictionary[key]
	if !valueExists {
		return false
	}

	delete(o.dictionary, key)
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
func (o *OrderedMap[K, V]) Clone() (cloned *OrderedMap[K, V]) {
	cloned = New[K, V]()
	o.ForEach(func(key K, value V) bool {
		cloned.Set(key, value)

		return true
	})

	return
}

// Encode returns a serialized byte slice of the object.
func (o *OrderedMap[K, V]) Encode() ([]byte, error) {
	seri := serializer.NewSerializer()

	seri.WriteNum(uint32(o.Size()), func(err error) error {
		return xerrors.Errorf("failed to write OrderedMap size to serializer: %w", err)
	})

	o.ForEach(func(key K, val V) bool {
		keyBytes, err := serix.DefaultAPI.Encode(context.Background(), key)
		if err != nil {
			seri.AbortIf(func(_ error) error {
				return xerrors.Errorf("failed to encode OrderedMap key: %w", err)
			})
		}
		seri.WriteBytes(keyBytes, func(err error) error {
			return xerrors.Errorf("failed to write OrderedMap key to serializer: %w", err)
		})

		valBytes, err := serix.DefaultAPI.Encode(context.Background(), val)
		if err != nil {
			seri.AbortIf(func(_ error) error {
				return xerrors.Errorf("failed to serialize OrderedMap value: %w", err)
			})
		}
		seri.WriteBytes(valBytes, func(err error) error {
			return xerrors.Errorf("failed to write OrderedMap value to serializer: %w", err)
		})

		return true
	})

	return seri.Serialize()
}

// Decode deserializes bytes into a valid object.
func (o *OrderedMap[K, V]) Decode(b []byte) (bytesRead int, err error) {
	o.Initialize()

	var mapSize uint32
	bytesReadSize, err := serix.DefaultAPI.Decode(context.Background(), b[bytesRead:], &mapSize)
	if err != nil {
		return 0, err
	}
	bytesRead += bytesReadSize

	for i := uint32(0); i < mapSize; i++ {
		var key K
		bytesReadKey, err := serix.DefaultAPI.Decode(context.Background(), b[bytesRead:], &key)
		if err != nil {
			return 0, err
		}
		bytesRead += bytesReadKey

		var value V
		bytesReadValue, err := serix.DefaultAPI.Decode(context.Background(), b[bytesRead:], &value)
		if err != nil {
			return 0, err
		}
		bytesRead += bytesReadValue

		o.Set(key, value)
	}

	return bytesRead, nil
}

package orderedmap

import (
	"context"

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/datastructure/orderedmap"
	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/hive.go/serix"
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

// Clone returns a copy of the orderedMap.
func (orderedMap *OrderedMap[K, V]) Clone() (cloned *OrderedMap[K, V]) {
	cloned = New[K, V]()
	orderedMap.OrderedMap.ForEach(func(key, value interface{}) bool {
		cloned.Set(key.(K), value.(V))
		return true
	})
	return
}

// Encode returns a serialized byte slice of the object.
func (orderedMap *OrderedMap[K, V]) Encode() ([]byte, error) {
	seri := serializer.NewSerializer()

	seri.WriteNum(uint32(orderedMap.Size()), func(err error) error {
		return errors.Wrap(err, "failed to write OrderedMap size to serializer")
	})

	orderedMap.ForEach(func(key K, val V) bool {
		keyBytes, err := serix.DefaultAPI.Encode(context.Background(), key)
		if err != nil {
			seri.AbortIf(func(err error) error {
				return errors.Wrap(err, "encode OrderedMap key")
			})
		}
		seri.WriteBytes(keyBytes, func(err error) error {
			return errors.Wrap(err, "failed to write OrderedMap key to serializer")
		})

		valBytes, err := serix.DefaultAPI.Encode(context.Background(), val)
		seri.AbortIf(func(_ error) error {
			return errors.Wrap(err, "failed to serialize OrderedMap value")
		})
		seri.WriteBytes(valBytes, func(err error) error {
			return errors.Wrap(err, "failed to write OrderedMap value to serializer")
		})
		return true
	})
	return seri.Serialize()
}

// Decode deserializes bytes into a valid object.
func (orderedMap *OrderedMap[K, V]) Decode(b []byte) (bytesRead int, err error) {
	orderedMap.OrderedMap = orderedmap.New()
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

		orderedMap.Set(key, value)
	}

	return bytesRead, nil
}

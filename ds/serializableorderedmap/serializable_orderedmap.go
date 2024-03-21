package serializableorderedmap

import (
	"context"

	"github.com/iotaledger/hive.go/ds/orderedmap"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

// SerializableOrderedMap provides a concurrent-safe ordered map that is serializable.
type SerializableOrderedMap[K comparable, V any] struct {
	*orderedmap.OrderedMap[K, V]
}

// New returns a new *SerializableOrderedMap.
func New[K comparable, V any]() *SerializableOrderedMap[K, V] {
	return &SerializableOrderedMap[K, V]{
		OrderedMap: orderedmap.New[K, V](),
	}
}

// Encode returns a serialized byte slice of the object.
func (o *SerializableOrderedMap[K, V]) Encode(api *serix.API) ([]byte, error) {
	seri := serializer.NewSerializer()

	seri.WriteNum(uint32(o.Size()), func(err error) error {
		return ierrors.Wrap(err, "failed to write SerializableOrderedMap size to serializer")
	})

	o.ForEach(func(key K, val V) bool {
		keyBytes, err := api.Encode(context.Background(), key)
		if err != nil {
			seri.AbortIf(func(_ error) error {
				return ierrors.Wrap(err, "failed to encode SerializableOrderedMap key")
			})
		}
		seri.WriteBytes(keyBytes, func(err error) error {
			return ierrors.Wrap(err, "failed to write SerializableOrderedMap key to serializer")
		})

		valBytes, err := api.Encode(context.Background(), val)
		if err != nil {
			seri.AbortIf(func(_ error) error {
				return ierrors.Wrap(err, "failed to serialize SerializableOrderedMap value")
			})
		}
		seri.WriteBytes(valBytes, func(err error) error {
			return ierrors.Wrap(err, "failed to write SerializableOrderedMap value to serializer")
		})

		return true
	})

	return seri.Serialize()
}

// Decode deserializes bytes into a valid object.
func (o *SerializableOrderedMap[K, V]) Decode(api *serix.API, b []byte) (bytesRead int, err error) {
	var mapSize uint32
	bytesReadSize, err := api.Decode(context.Background(), b[bytesRead:], &mapSize)
	if err != nil {
		return 0, err
	}
	bytesRead += bytesReadSize

	for range mapSize {
		var key K
		bytesReadKey, err := api.Decode(context.Background(), b[bytesRead:], &key)
		if err != nil {
			return 0, err
		}
		bytesRead += bytesReadKey

		var value V
		bytesReadValue, err := api.Decode(context.Background(), b[bytesRead:], &value)
		if err != nil {
			return 0, err
		}
		bytesRead += bytesReadValue

		o.Set(key, value)
	}

	return bytesRead, nil
}

package customtypes

import (
	"context"

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/datastructure/genericcomparator"
	"github.com/iotaledger/hive.go/generics/thresholdmap"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serix"
)

// region SerializableThresholdMap /////////////////////////////////////////////////////////////////////////////////////////

// SerializableThresholdMap is a wrapper around generic ThresholdMap which implements Serializable interface.
type SerializableThresholdMap[K comparable, V any] struct {
	*thresholdmap.ThresholdMap[K, V]
}

func NewSerializableThresholdMap[K comparable, V any](mode thresholdmap.Mode, optionalComparator ...genericcomparator.Type) *SerializableThresholdMap[K, V] {
	return &SerializableThresholdMap[K, V]{thresholdmap.New[K, V](mode, optionalComparator...)}
}

// Encode returns a serialized byte slice of the object.
func (s *SerializableThresholdMap[K, V]) Encode() ([]byte, error) {
	seri := serializer.NewSerializer()

	seri.WriteNum(uint32(s.ThresholdMap.Size()), func(err error) error {
		return errors.Wrap(err, "failed to write ThresholdMap size to serializer")
	})

	s.ThresholdMap.ForEach(func(elem *thresholdmap.Element[K, V]) bool {
		keyBytes, err := serix.DefaultAPI.Encode(context.Background(), elem.Key())
		seri.AbortIf(func(_ error) error {
			return errors.Wrap(err, "encode ThresholdMap key")
		})
		seri.WriteBytes(keyBytes, func(err error) error {
			return errors.Wrap(err, "failed to ThresholdMap key to serializer")
		})

		valBytes, err := serix.DefaultAPI.Encode(context.Background(), elem.Value())
		seri.AbortIf(func(_ error) error {
			return errors.Wrap(err, "failed to serialize ThresholdMap value")
		})
		seri.WriteBytes(valBytes, func(err error) error {
			return errors.Wrap(err, "failed to ThresholdMap value to serializer")
		})
		return true
	})
	return seri.Serialize()
}

// Decode deserializes bytes into a valid object.
func (s *SerializableThresholdMap[K, V]) Decode(b []byte) (bytesRead int, err error) {
	//var mapSize uint32
	//bytesReadSize, err := serix.DefaultAPI.Decode(context.Background(), b[bytesRead:], &mapSize)
	//if err != nil {
	//	return 0, err
	//}
	//bytesRead += bytesReadSize
	//
	//for i := uint32(0); i < mapSize; i++ {
	//	var key K
	//	bytesReadKey, err := serix.DefaultAPI.Decode(context.Background(), b[bytesRead:], &key)
	//	if err != nil {
	//		return 0, err
	//	}
	//	bytesRead += bytesReadKey
	//
	//	var value V
	//	bytesReadValue, err := serix.DefaultAPI.Decode(context.Background(), b[bytesRead:], &value)
	//	if err != nil {
	//		return 0, err
	//	}
	//	bytesRead += bytesReadValue
	//
	//	s.Set(key, value)
	//}

	return bytesRead, nil
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

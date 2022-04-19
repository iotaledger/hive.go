package customtypes

import (
	"context"

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/generics/set"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serix"
)

// region Voters ///////////////////////////////////////////////////////////////////////////////////////////////////////

// SerializableSet is a wrapper around generic Set which implements Serializable interface.
type SerializableSet[T comparable] struct {
	set.Set[T]
}

// NewSerializableSet is the constructor SerializableSet.
func NewSerializableSet[T comparable](threadSafe ...bool) (voters SerializableSet[T]) {
	return SerializableSet[T]{set.New[T](threadSafe...)}
}

// Encode returns a serialized byte slice of the object.
func (v SerializableSet[T]) Encode() ([]byte, error) {
	seri := serializer.NewSerializer()

	seri.WriteNum(uint32(v.Size()), func(err error) error {
		return errors.Wrap(err, "failed to write set size to serializer")
	})

	v.ForEach(func(elem T) {
		bytes, err := serix.DefaultAPI.Encode(context.Background(), elem)
		if err != nil {
			seri.AbortIf(func(err error) error {
				return errors.Wrap(err, "failed to serialize element of a set")
			})
		}
		seri.WriteBytes(bytes, func(err error) error {
			return errors.Wrap(err, "failed to write elem to serializer")
		})
	})
	return seri.Serialize()
}

// Decode deserializes bytes into a valid object.
func (v SerializableSet[T]) Decode(b []byte) (bytesRead int, err error) {
	var elemCount uint32
	bytesRead, err = serix.DefaultAPI.Decode(context.Background(), b, &elemCount)
	if err != nil {
		return 0, err
	}

	for i := uint32(0); i < elemCount; i++ {
		var elem T
		bytesReadVoter, err := serix.DefaultAPI.Decode(context.Background(), b[bytesRead:], &elem)
		if err != nil {
			return 0, err
		}
		bytesRead += bytesReadVoter
		v.Set.Add(elem)
	}
	return bytesRead, nil
}

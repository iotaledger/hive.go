package set

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/iotaledger/hive.go/datastructure/set"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serix"
)

// type genericSet[T comparable] struct { implements a generic wrapper for a non-generic Set.
type genericSet[T comparable] struct {
	set.Set
}

// newGenericSetWrapper returns a new generic Set.
func newGenericSet[T comparable](s set.Set) *genericSet[T] {
	return &genericSet[T]{
		Set: s,
	}
}

// Add adds a new element to the Set and returns true if the element was not present in the set before.
func (set *genericSet[T]) Add(element T) bool {
	return set.Set.Add(element)
}

// Delete removes the element from the Set and returns true if it did exist.
func (set *genericSet[T]) Delete(element T) bool {
	return set.Set.Delete(element)
}

// Has returns true if the element exists in the Set.
func (set *genericSet[T]) Has(element T) bool {
	return set.Set.Has(element)
}

// ForEach iterates through the set and calls the callback for every element.
func (set *genericSet[T]) ForEach(callback func(element T)) {
	set.Set.ForEach(func(element interface{}) {
		callback(element.(T))
	})
}

// Encode returns a serialized byte slice of the object.
func (set *genericSet[T]) Encode() ([]byte, error) {
	seri := serializer.NewSerializer()

	seri.WriteNum(uint32(set.Size()), func(err error) error {
		return errors.Wrap(err, "failed to write set size to serializer")
	})

	set.ForEach(func(elem T) {
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
func (set *genericSet[T]) Decode(b []byte) (bytesRead int, err error) {
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
		set.Set.Add(elem)
	}
	return bytesRead, nil
}

// code contract - make sure the type implements the interface
var _ Set[int] = &genericSet[int]{}

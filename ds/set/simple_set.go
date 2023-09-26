package set

import (
	"context"

	"github.com/pkg/errors"

	"github.com/izuc/zipp.foundation/ds/types"
	"github.com/izuc/zipp.foundation/serializer/v2"
	"github.com/izuc/zipp.foundation/serializer/v2/serix"
)

// simpleSet implements a non-thread safe Set.
type simpleSet[T comparable] struct {
	elements map[T]types.Empty
}

// newSimpleSet returns a new non-thread safe Set.
func newSimpleSet[T comparable]() *simpleSet[T] {
	s := new(simpleSet[T])
	s.initialize()
	return s
}

func (s *simpleSet[T]) initialize() {
	s.elements = make(map[T]types.Empty)
}

// Add adds a new element to the Set and returns true if the element was not present in the set before.
func (s *simpleSet[T]) Add(element T) bool {
	if _, elementExists := s.elements[element]; elementExists {
		return false
	}

	s.elements[element] = types.Void

	return true
}

// Delete removes the element from the Set and returns true if it did exist.
func (s *simpleSet[T]) Delete(element T) bool {
	_, elementExists := s.elements[element]
	if elementExists {
		delete(s.elements, element)
	}

	return elementExists
}

// Has returns true if the element exists in the Set.
func (s *simpleSet[T]) Has(element T) bool {
	_, elementExists := s.elements[element]

	return elementExists
}

// ForEach iterates through the set and calls the callback for every element.
func (s *simpleSet[T]) ForEach(callback func(element T)) {
	for element := range s.elements {
		callback(element)
	}
}

// Clear removes all elements from the Set.
func (s *simpleSet[T]) Clear() {
	s.elements = make(map[T]types.Empty)
}

// Size returns the size of the Set.
func (s *simpleSet[T]) Size() int {
	return len(s.elements)
}

// Encode returns a serialized byte slice of the object.
func (s *simpleSet[T]) Encode() ([]byte, error) {
	seri := serializer.NewSerializer()

	seri.WriteNum(uint32(s.Size()), func(err error) error {
		return errors.Wrap(err, "failed to write set size to serializer")
	})

	s.ForEach(func(elem T) {
		bytes, err := serix.DefaultAPI.Encode(context.Background(), elem)
		if err != nil {
			seri.AbortIf(func(_ error) error {
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
func (s *simpleSet[T]) Decode(b []byte) (bytesRead int, err error) {
	s.initialize()

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
		s.Add(elem)
	}

	return bytesRead, nil
}

// code contract - make sure the type implements the interface.
var _ Set[int] = &simpleSet[int]{}

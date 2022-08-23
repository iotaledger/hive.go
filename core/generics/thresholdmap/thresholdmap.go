//nolint:golint // golint throws false positives with generics here
package thresholdmap

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/iotaledger/hive.go/core/datastructure/thresholdmap"
	"github.com/iotaledger/hive.go/core/generics/constraints"
	"github.com/iotaledger/hive.go/core/generics/lo"
	"github.com/iotaledger/hive.go/core/serix"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// region Mode /////////////////////////////////////////////////////////////////////////////////////////////////////////

// Mode encodes different modes of function for the ThresholdMap that specifies if the defines keys act as upper or
// lower thresholds.
type Mode bool

const (
	// LowerThresholdMode interprets the keys of the ThresholdMap as lower thresholds which means that querying the map
	// will return the value of the largest node whose key is <= than the queried value.
	LowerThresholdMode = true

	// UpperThresholdMode interprets the keys of the ThresholdMap as upper thresholds which means that querying the map
	// will return the value of the smallest node whose key is >= than the queried value.
	UpperThresholdMode = false
)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ThresholdMap /////////////////////////////////////////////////////////////////////////////////////////////////

// ThresholdMap is a data structure that allows to map keys bigger or lower than a certain threshold to a given value.
type ThresholdMap[K constraints.Ordered, V any] struct {
	thresholdmap.ThresholdMap
}

// New returns a ThresholdMap that operates in the given Mode and that can also receive an optional comparator function
// to support custom key types.
func New[K constraints.Ordered, V any](mode Mode) *ThresholdMap[K, V] {
	return &ThresholdMap[K, V]{
		ThresholdMap: *thresholdmap.New(thresholdmap.Mode(mode), func(a interface{}, b interface{}) int {
			return lo.Comparator(a.(K), b.(K))
		}),
	}
}

// Set adds a new threshold that maps all keys >= or <= (depending on the Mode) the value given by key to a certain
// value.
func (t *ThresholdMap[K, V]) Set(key K, value V) {
	t.ThresholdMap.Set(key, value)
}

// Get returns the value of the next higher or lower existing threshold (depending on the mode) and a flag that
// indicates if there is a threshold that covers the given value.
func (t *ThresholdMap[K, V]) Get(key K) (value V, exists bool) {
	v, exists := t.ThresholdMap.Get(key)
	if exists {
		value = v.(V)
	}

	return
}

// Floor returns the largest key that is <= the given key, it's value and a boolean flag indicating if it exists.
func (t *ThresholdMap[K, V]) Floor(key K) (floorKey K, floorValue V, exists bool) {
	floor, value, exists := t.ThresholdMap.Floor(key)
	if exists {
		floorKey = floor.(K)
		floorValue = value.(V)
	}

	return
}

// Ceiling returns the smallest key that is >= the given key, it's value and a boolean flag indicating if it exists.
func (t *ThresholdMap[K, V]) Ceiling(key K) (floorKey K, floorValue V, exists bool) {
	ceil, value, exists := t.ThresholdMap.Ceiling(key)
	if exists {
		floorKey = ceil.(K)
		floorValue = value.(V)
	}

	return
}

// Delete removes a threshold from the map.
func (t *ThresholdMap[K, V]) Delete(key K) (element *Element[K, V], success bool) {
	elem, success := t.ThresholdMap.Delete(key)
	if success {
		element = t.wrapNode(elem)
	}

	return
}

// Keys returns a list of thresholds that have been set in the map.
func (t *ThresholdMap[K, V]) Keys() []K {
	rawKeys := t.ThresholdMap.Keys()
	result := make([]K, len(rawKeys))
	for i, v := range rawKeys {
		result[i] = v.(K)
	}

	return result
}

// Values returns a list of values that are associated to the thresholds in the map.
func (t *ThresholdMap[K, V]) Values() []V {
	rawVals := t.ThresholdMap.Values()
	result := make([]V, len(rawVals))
	for i, v := range rawVals {
		result[i] = v.(V)
	}

	return result
}

// GetElement returns the Element that is used to store the next higher or lower threshold (depending on the mode)
// belonging to the given key (or nil if none exists).
func (t *ThresholdMap[K, V]) GetElement(key K) *Element[K, V] {
	elem := t.ThresholdMap.GetElement(key)

	return t.wrapNode(elem)
}

// MinElement returns the smallest threshold in the map (or nil if the map is empty).
func (t *ThresholdMap[K, V]) MinElement() *Element[K, V] {
	return t.wrapNode(t.ThresholdMap.MinElement())
}

// MaxElement returns the largest threshold in the map (or nil if the map is empty).
func (t *ThresholdMap[K, V]) MaxElement() *Element[K, V] {
	return t.wrapNode(t.ThresholdMap.MaxElement())
}

// DeleteElement removes the given Element from the map.
func (t *ThresholdMap[K, V]) DeleteElement(element *Element[K, V]) {
	t.ThresholdMap.DeleteElement(element.Element)

}

// ForEach provides a callback based iterator that iterates through all Elements in the map.
func (t *ThresholdMap[K, V]) ForEach(iterator func(node *Element[K, V]) bool) {
	t.ThresholdMap.ForEach(func(node *thresholdmap.Element) bool {
		return iterator(t.wrapNode(node))
	})
}

// Iterator returns an Iterator object that can be used to manually iterate through the Elements in the map. It accepts
// an optional starting Element where the iteration begins.
func (t *ThresholdMap[K, V]) Iterator(optionalStartingNode ...*Element[K, V]) *Iterator[K, V] {
	if len(optionalStartingNode) > 0 {
		return NewIterator[K, V](t.ThresholdMap.Iterator(optionalStartingNode[0].Element))
	}

	return NewIterator[K, V](t.ThresholdMap.Iterator())
}

// wrapNode is an internal utility function that wraps the Node of the underlying RedBlackTree with a map Element.
func (t *ThresholdMap[K, V]) wrapNode(node *thresholdmap.Element) (element *Element[K, V]) {
	if node == nil {
		return
	}

	return &Element[K, V]{node}
}

// Encode returns a serialized byte slice of the object.
func (t *ThresholdMap[K, V]) Encode() ([]byte, error) {
	t.RLock()
	defer t.RUnlock()

	seri := serializer.NewSerializer()

	seri.WriteBool(bool(t.Mode()), func(err error) error {
		return errors.Errorf("failed to write mode: %w", err)
	})

	seri.WriteNum(uint32(t.ThresholdMap.Size()), func(err error) error {
		return errors.Wrap(err, "failed to write ThresholdMap size to serializer")
	})

	t.ForEach(func(elem *Element[K, V]) bool {
		keyBytes, err := serix.DefaultAPI.Encode(context.Background(), elem.Key())
		seri.AbortIf(func(_ error) error {
			return errors.Wrap(err, "encode ThresholdMap key")
		})
		seri.WriteBytes(keyBytes, func(err error) error {
			return errors.Wrap(err, "failed to write ThresholdMap key to serializer")
		})

		valBytes, err := serix.DefaultAPI.Encode(context.Background(), elem.Value())
		seri.AbortIf(func(_ error) error {
			return errors.Wrap(err, "failed to serialize ThresholdMap value")
		})
		seri.WriteBytes(valBytes, func(err error) error {
			return errors.Wrap(err, "failed to write ThresholdMap value to serializer")
		})

		return true
	})

	return seri.Serialize()
}

// Decode deserializes bytes into a valid object.
func (t *ThresholdMap[K, V]) Decode(b []byte) (bytesRead int, err error) {
	var mode thresholdmap.Mode
	bytesRead, err = serix.DefaultAPI.Decode(context.Background(), b, &mode, serix.WithValidation())
	if err != nil {
		return bytesRead, errors.Errorf("failed to decode mode: %w", err)
	}

	t.Init(mode, func(a interface{}, b interface{}) int { return lo.Comparator(a.(K), b.(K)) })

	var mapSize uint32
	bytesReadMapSize, err := serix.DefaultAPI.Decode(context.Background(), b[bytesRead:], &mapSize)
	bytesRead += bytesReadMapSize
	if err != nil {
		return 0, err
	}

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

		t.Set(key, value)
	}

	return bytesRead, nil
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Element //////////////////////////////////////////////////////////////////////////////////////////////////////

// Element is a wrapper for the Node used in the underlying red-black RedBlackTree.
type Element[K any, V any] struct {
	*thresholdmap.Element
}

// Key returns the Key of the Element.
func (e *Element[K, V]) Key() K {
	return e.Node.Key.(K)
}

// Value returns the Value of the Element.
func (e *Element[K, V]) Value() V {
	return e.Node.Value.(V)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Iterator /////////////////////////////////////////////////////////////////////////////////////////////////////

// Iterator is an object that allows to iterate over the ThresholdMap by providing methods to walk through the map in a
// deterministic order.
type Iterator[K any, V any] struct {
	*thresholdmap.Iterator
}

// NewIterator is the constructor of the Iterator that takes the starting Element as its parameter.
func NewIterator[K any, V any](iterator *thresholdmap.Iterator) *Iterator[K, V] {
	return &Iterator[K, V]{
		Iterator: iterator,
	}
}

// Next returns the next Element in the Iterator and advances the internal pointer. The method panics if there is no
// next Element that can be retrieved (always use HasNext to check if another Element can be requested).
func (i *Iterator[K, V]) Next() *Element[K, V] {
	return i.wrapNode(i.Iterator.Next())

}

// Prev returns the previous Element in the Iterator and moves back the internal pointer. The method panics if there is
// no previous Element that can be retrieved (always use HasPrev to check if another Element can be requested).
func (i *Iterator[K, V]) Prev() *Element[K, V] {
	return i.wrapNode(i.Iterator.Prev())

}

// wrapNode is an internal utility function that wraps the Node of the underlying RedBlackTree with a map Element.
func (i *Iterator[K, V]) wrapNode(node *thresholdmap.Element) (element *Element[K, V]) {
	if node == nil {
		return
	}

	return &Element[K, V]{node}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

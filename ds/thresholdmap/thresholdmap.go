package thresholdmap

import (
	"context"
	"sync"

	"github.com/emirpasic/gods/trees/redblacktree"
	"golang.org/x/xerrors"

	"github.com/izuc/zipp.foundation/constraints"
	"github.com/izuc/zipp.foundation/lo"
	"github.com/izuc/zipp.foundation/serializer/v2"
	"github.com/izuc/zipp.foundation/serializer/v2/serix"
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
	mode Mode
	tree *redblacktree.Tree

	sync.RWMutex
}

// New returns a ThresholdMap that operates in the given Mode.
func New[K constraints.Ordered, V any](mode Mode) *ThresholdMap[K, V] {
	return new(ThresholdMap[K, V]).Init(mode)
}

// Init initializes the ThresholdMap with the given Mode and comparator function.
func (t *ThresholdMap[K, V]) Init(mode Mode) *ThresholdMap[K, V] {
	t.Lock()
	defer t.Unlock()
	if t.tree != nil {
		panic("ThresholdMap has already been initialized before")
	}

	t.mode = mode
	t.tree = redblacktree.NewWith(func(a interface{}, b interface{}) int {
		return lo.Comparator(a.(K), b.(K))
	})

	return t
}

// Mode returns the mode of this ThresholdMap.
func (t *ThresholdMap[K, V]) Mode() Mode {
	t.RLock()
	defer t.RUnlock()

	return t.mode
}

// Set adds a new threshold that maps all keys >= or <= (depending on the Mode) the value given by key to a certain
// value.
func (t *ThresholdMap[K, V]) Set(key K, value V) {
	t.Lock()
	defer t.Unlock()
	t.tree.Put(key, value)
}

// Get returns the value of the next higher or lower existing threshold (depending on the mode) and a flag that
// indicates if there is a threshold that covers the given value.
func (t *ThresholdMap[K, V]) Get(key K) (value V, exists bool) {
	t.RLock()
	defer t.RUnlock()
	var foundNode *redblacktree.Node
	switch t.mode {
	case UpperThresholdMode:
		foundNode, exists = t.tree.Ceiling(key)
	case LowerThresholdMode:
		foundNode, exists = t.tree.Floor(key)
	default:
		panic("unsupported mode")
	}

	if exists {
		value = foundNode.Value.(V)
	}

	return
}

// Floor returns the largest key that is <= the given key, it's value and a boolean flag indicating if it exists.
func (t *ThresholdMap[K, V]) Floor(key K) (floorKey K, floorValue V, exists bool) {
	t.RLock()
	defer t.RUnlock()
	if node, exists := t.tree.Floor(key); exists {
		return node.Key.(K), node.Value.(V), true
	}

	return floorKey, floorValue, false
}

// Ceiling returns the smallest key that is >= the given key, it's value and a boolean flag indicating if it exists.
func (t *ThresholdMap[K, V]) Ceiling(key K) (ceilingKey K, ceilingValue V, exists bool) {
	t.RLock()
	defer t.RUnlock()
	if node, exists := t.tree.Ceiling(key); exists {
		return node.Key.(K), node.Value.(V), true
	}

	return ceilingKey, ceilingValue, false
}

// Delete removes a threshold from the map.
func (t *ThresholdMap[K, V]) Delete(key K) (element *Element[K, V], success bool) {
	t.Lock()
	defer t.Unlock()
	node := t.lookup(key)
	if node != nil {
		t.tree.Remove(key)
	}

	return t.wrapNode(node), node != nil
}

func (t *ThresholdMap[K, V]) lookup(key interface{}) *redblacktree.Node {
	node := t.tree.Root
	for node != nil {
		compare := t.tree.Comparator(key, node.Key)
		switch {
		case compare == 0:
			return node
		case compare < 0:
			node = node.Left
		case compare > 0:
			node = node.Right
		}
	}

	return nil
}

// Keys returns a list of thresholds that have been set in the map.
func (t *ThresholdMap[K, V]) Keys() []K {
	t.RLock()
	defer t.RUnlock()

	return lo.Map(t.tree.Keys(), func(key interface{}) K {
		return key.(K)
	})
}

// Values returns a list of values that are associated to the thresholds in the map.
func (t *ThresholdMap[K, V]) Values() []V {
	t.RLock()
	defer t.RUnlock()

	return lo.Map(t.tree.Values(), func(value interface{}) V {
		return value.(V)
	})
}

// GetElement returns the Element that is used to store the next higher or lower threshold (depending on the mode)
// belonging to the given key (or nil if none exists).
func (t *ThresholdMap[K, V]) GetElement(key K) *Element[K, V] {
	t.RLock()
	defer t.RUnlock()
	switch t.mode {
	case UpperThresholdMode:
		ceiling, _ := t.tree.Ceiling(key)

		return t.wrapNode(ceiling)
	default:
		floor, _ := t.tree.Floor(key)

		return t.wrapNode(floor)
	}
}

// MinElement returns the smallest threshold in the map (or nil if the map is empty).
func (t *ThresholdMap[K, V]) MinElement() *Element[K, V] {
	t.RLock()
	defer t.RUnlock()

	return t.wrapNode(t.tree.Left())
}

// MaxElement returns the largest threshold in the map (or nil if the map is empty).
func (t *ThresholdMap[K, V]) MaxElement() *Element[K, V] {
	t.RLock()
	defer t.RUnlock()

	return t.wrapNode(t.tree.Right())
}

// DeleteElement removes the given Element from the map.
func (t *ThresholdMap[K, V]) DeleteElement(element *Element[K, V]) {
	t.Lock()
	defer t.Unlock()
	if element == nil {
		return
	}

	t.tree.Remove(element.Node)
}

// ForEach provides a callback based iterator that iterates through all Elements in the map.
func (t *ThresholdMap[K, V]) ForEach(iterator func(node *Element[K, V]) bool) {
	t.RLock()
	defer t.RUnlock()
	for it := t.tree.Iterator(); it.Next(); {
		if !iterator(t.wrapNode(&redblacktree.Node{Key: it.Key(), Value: it.Value()})) {
			break
		}
	}
}

// Iterator returns an Iterator object that can be used to manually iterate through the Elements in the map. It accepts
// an optional starting Element where the iteration begins.
func (t *ThresholdMap[K, V]) Iterator(optionalStartingNode ...*Element[K, V]) *Iterator[K, V] {
	t.RLock()
	defer t.RUnlock()
	if len(optionalStartingNode) >= 1 {
		return NewIterator(optionalStartingNode[0])
	}

	return NewIterator(t.wrapNode(t.tree.Left()))
}

// Size returns the amount of thresholds that are stored in the map.
func (t *ThresholdMap[K, V]) Size() int {
	t.RLock()
	defer t.RUnlock()

	return t.tree.Size()
}

// Empty returns true of the map has no thresholds.
func (t *ThresholdMap[K, V]) Empty() bool {
	t.RLock()
	defer t.RUnlock()

	return t.tree.Empty()
}

// Clear removes all Elements from the map.
func (t *ThresholdMap[K, V]) Clear() {
	t.Lock()
	defer t.Unlock()
	t.tree.Clear()
}

// wrapNode is an internal utility function that wraps the Node of the underlying RedBlackTree with a map Element.
func (t *ThresholdMap[K, V]) wrapNode(node *redblacktree.Node) (element *Element[K, V]) {
	if node == nil {
		return
	}

	return &Element[K, V]{node}
}

// Encode returns a serialized byte slice of the object.
func (t ThresholdMap[K, V]) Encode() ([]byte, error) {
	t.RLock()
	defer t.RUnlock()

	seri := serializer.NewSerializer()

	seri.WriteBool(bool(t.mode), func(err error) error {
		return xerrors.Errorf("failed to write mode: %w", err)
	})

	seri.WriteNum(uint32(t.tree.Size()), func(err error) error {
		return xerrors.Errorf("failed to write ThresholdMap size to serializer: %w", err)
	})

	t.ForEach(func(elem *Element[K, V]) bool {
		keyBytes, err := serix.DefaultAPI.Encode(context.Background(), elem.Key())
		if err != nil {
			seri.AbortIf(func(_ error) error {
				return xerrors.Errorf("failed to encode ThresholdMap key: %w", err)
			})
		}
		seri.WriteBytes(keyBytes, func(err error) error {
			return xerrors.Errorf("failed to write ThresholdMap key to serializer: %w", err)
		})

		valBytes, err := serix.DefaultAPI.Encode(context.Background(), elem.Value())
		if err != nil {
			seri.AbortIf(func(_ error) error {
				return xerrors.Errorf("failed to serialize ThresholdMap value: %w", err)
			})
		}
		seri.WriteBytes(valBytes, func(err error) error {
			return xerrors.Errorf("failed to write ThresholdMap value to serializer", err)
		})

		return true
	})

	return seri.Serialize()
}

// Decode deserializes bytes into a valid object.
func (t *ThresholdMap[K, V]) Decode(b []byte) (bytesRead int, err error) {
	var mode Mode
	bytesRead, err = serix.DefaultAPI.Decode(context.Background(), b, &mode, serix.WithValidation())
	if err != nil {
		return bytesRead, xerrors.Errorf("failed to decode mode: %w", err)
	}

	t.Init(mode)

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

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

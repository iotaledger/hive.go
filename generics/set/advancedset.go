package set

import (
	"github.com/cockroachdb/errors"

	"github.com/iotaledger/hive.go/generics/orderedmap"
	"github.com/iotaledger/hive.go/generics/walker"
	"github.com/iotaledger/hive.go/types"
)

type AdvancedSet[T comparable] struct {
	orderedmap.OrderedMap[T, types.Empty] `serix:"0"`
}

func NewAdvancedSet[T comparable](elements ...T) (new *AdvancedSet[T]) {
	new = &AdvancedSet[T]{*orderedmap.New[T, types.Empty]()}
	for _, element := range elements {
		new.Set(element, types.Void)
	}

	return new
}

func (t *AdvancedSet[T]) IsEmpty() (empty bool) {
	return t.OrderedMap.Size() == 0
}

func (t *AdvancedSet[T]) Add(element T) (added bool) {
	return t.Set(element, types.Void)
}

func (t *AdvancedSet[T]) AddAll(elements *AdvancedSet[T]) (added bool) {
	_ = elements.ForEach(func(element T) (err error) {
		added = t.Set(element, types.Void) || added
		return nil
	})

	return added
}

func (t *AdvancedSet[T]) DeleteAll(other *AdvancedSet[T]) (removedElements *AdvancedSet[T]) {
	removedElements = NewAdvancedSet[T]()
	_ = other.ForEach(func(element T) (err error) {
		if t.Delete(element) {
			removedElements.Add(element)
		}
		return nil
	})

	return removedElements
}

func (t *AdvancedSet[T]) Delete(element T) (added bool) {
	return t.OrderedMap.Delete(element)
}

func (t *AdvancedSet[T]) ForEach(callback func(element T) (err error)) (err error) {
	t.OrderedMap.ForEach(func(element T, _ types.Empty) bool {
		if err = callback(element); err != nil {
			return false
		}

		return true
	})

	return err
}

func (t *AdvancedSet[T]) Intersect(other *AdvancedSet[T]) (intersection *AdvancedSet[T]) {
	return t.Filter(other.Has)
}

func (t *AdvancedSet[T]) Filter(predicate func(element T) bool) (filtered *AdvancedSet[T]) {
	filtered = NewAdvancedSet[T]()
	_ = t.ForEach(func(element T) (err error) {
		if predicate(element) {
			filtered.Add(element)
		}

		return nil
	})

	return filtered
}

func (t *AdvancedSet[T]) Equal(other *AdvancedSet[T]) (equal bool) {
	if other.Size() != t.Size() {
		return false
	}

	return other.ForEach(func(element T) (err error) {
		if !t.Has(element) {
			return errors.New("abort")
		}

		return nil
	}) == nil
}

func (t *AdvancedSet[T]) Is(element T) bool {
	return t.Size() == 1 && t.Has(element)
}

func (t *AdvancedSet[T]) Clone() (cloned *AdvancedSet[T]) {
	cloned = NewAdvancedSet[T]()
	cloned.AddAll(t)

	return cloned
}

func (t *AdvancedSet[T]) Slice() (slice []T) {
	slice = make([]T, 0)
	_ = t.ForEach(func(element T) error {
		slice = append(slice, element)
		return nil
	})

	return slice
}

func (t *AdvancedSet[T]) Iterator() *walker.Walker[T] {
	return walker.New[T](false).PushAll(t.Slice()...)
}

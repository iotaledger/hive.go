package advancedset

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/izuc/zipp.foundation/ds/orderedmap"
	"github.com/izuc/zipp.foundation/ds/types"
	"github.com/izuc/zipp.foundation/ds/walker"
	"github.com/izuc/zipp.foundation/lo"
)

// AdvancedSet is a set that offers advanced features.
type AdvancedSet[T comparable] struct {
	orderedmap.OrderedMap[T, types.Empty] `serix:"0"`
}

// New creates a new AdvancedSet with given elements.
func New[T comparable](elements ...T) *AdvancedSet[T] {
	a := &AdvancedSet[T]{*orderedmap.New[T, types.Empty]()}
	for _, element := range elements {
		a.Set(element, types.Void)
	}

	return a
}

// IsEmpty returns true if the set is empty.
func (t *AdvancedSet[T]) IsEmpty() (empty bool) {
	if t == nil {
		return true
	}

	return t.OrderedMap.Size() == 0
}

// Add adds a new element to the Set and returns true if the element was not present in the set before.
func (t *AdvancedSet[T]) Add(element T) (added bool) {
	return !lo.Return2(t.Set(element, types.Void))
}

// AddAll adds all elements to the AdvancedSet and returns true if any element has been added.
func (t *AdvancedSet[T]) AddAll(elements *AdvancedSet[T]) (added bool) {
	_ = elements.ForEach(func(element T) (err error) {
		added = !lo.Return2(t.Set(element, types.Void)) || added

		return nil
	})

	return added
}

// DeleteAll deletes given elements from the set.
func (t *AdvancedSet[T]) DeleteAll(other *AdvancedSet[T]) (removedElements *AdvancedSet[T]) {
	removedElements = New[T]()
	_ = other.ForEach(func(element T) (err error) {
		if t.Delete(element) {
			removedElements.Add(element)
		}

		return nil
	})

	return removedElements
}

// Delete removes the element from the Set and returns true if it did exist.
func (t *AdvancedSet[T]) Delete(element T) (deleted bool) {
	return t.OrderedMap.Delete(element)
}

// HasAll returns true if all given elements are present in the set.
func (t *AdvancedSet[T]) HasAll(other *AdvancedSet[T]) (hasAll bool) {
	if t == nil {
		return false
	}

	return other.ForEach(func(element T) error {
		if !t.Has(element) {
			return errors.New("element not found")
		}

		return nil
	}) == nil
}

// ForEach iterates through the set and calls the callback for every element.
func (t *AdvancedSet[T]) ForEach(callback func(element T) error) (err error) {
	if t == nil {
		return nil
	}

	t.OrderedMap.ForEach(func(element T, _ types.Empty) bool {
		if err = callback(element); err != nil {
			return false
		}

		return true
	})

	return err
}

// Range iterates through the set and calls the callback for every element (without individual error handling).
func (t *AdvancedSet[T]) Range(callback func(element T)) {
	if t != nil {
		t.OrderedMap.ForEach(func(element T, _ types.Empty) bool {
			callback(element)

			return true
		})
	}
}

// Intersect returns a new set that contains all elements that are present in both sets.
func (t *AdvancedSet[T]) Intersect(other *AdvancedSet[T]) (intersection *AdvancedSet[T]) {
	return t.Filter(other.Has)
}

// Filter returns a new set that contains all elements that satisfy the given predicate.
func (t *AdvancedSet[T]) Filter(predicate func(element T) bool) (filtered *AdvancedSet[T]) {
	filtered = New[T]()
	_ = t.ForEach(func(element T) (err error) {
		if predicate(element) {
			filtered.Add(element)
		}

		return nil
	})

	return filtered
}

// Equal returns true if both sets contain the same elements.
func (t *AdvancedSet[T]) Equal(other *AdvancedSet[T]) (equal bool) {
	return t == other || (t != nil && other != nil && t.Size() == other.Size() && t.HasAll(other))
}

// Is returns true if the given element is the only element in the set.
func (t *AdvancedSet[T]) Is(element T) bool {
	return t.Size() == 1 && t.Has(element)
}

// Clone returns a new set that contains the same elements as the original set.
func (t *AdvancedSet[T]) Clone() (cloned *AdvancedSet[T]) {
	cloned = New[T]()
	cloned.AddAll(t)

	return cloned
}

// Slice returns a slice of all elements in the set.
func (t *AdvancedSet[T]) Slice() (slice []T) {
	slice = make([]T, 0)

	if t != nil {
		_ = t.ForEach(func(element T) error {
			slice = append(slice, element)

			return nil
		})
	}

	return slice
}

// Iterator returns a new iterator for the set.
func (t *AdvancedSet[T]) Iterator() *walker.Walker[T] {
	return walker.New[T](false).PushAll(t.Slice()...)
}

// Intersect returns a new set that contains all elements that are present in both sets.
func (t *AdvancedSet[T]) String() (humanReadable string) {
	var elementType T
	elementTypeName := reflect.TypeOf(elementType).Name()

	elementStrings := make([]string, 0)
	_ = t.ForEach(func(element T) (err error) {
		elementStrings = append(elementStrings, strings.TrimRight(strings.ReplaceAll(fmt.Sprintf("%+v", element), elementTypeName+"(", ""), ")"))

		return nil
	})

	return fmt.Sprintf("%ss(%s)", elementTypeName, strings.Join(elementStrings, ", "))
}

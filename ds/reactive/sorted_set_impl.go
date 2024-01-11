package reactive

import (
	"cmp"

	"github.com/iotaledger/hive.go/ds"
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/runtime/syncutils"
)

// region sortedSet ////////////////////////////////////////////////////////////////////////////////////////////////////

// sortedSet is the default implementation of the SortedSet interface.
type sortedSet[ElementType comparable, WeightType cmp.Ordered] struct {
	// Set imports the methods of the Set interface.
	Set[ElementType]

	// elements is a map of all elements that are part of the set.
	elements *shrinkingmap.ShrinkingMap[ElementType, *sortedSetElement[ElementType, WeightType]]

	// sortedElements is a slice of all elements that are part of the set, sorted by their weight.
	sortedElements []*sortedSetElement[ElementType, WeightType]

	// heaviestElement is a reference to the element with the heaviest weight.
	heaviestElement Variable[ElementType]

	// lightestElement is a reference to the element with the lightest weight.
	lightestElement Variable[ElementType]

	// weightVariable is the function that is used to retrieve the weight of an element.
	weightVariable func(element ElementType) Variable[WeightType]

	// tieBreaker is the function that is used to break ties between elements with the same weight.
	tieBreaker func(left, right ElementType) int

	// mutex is used to synchronize access to the sortedElements slice.
	mutex syncutils.RWMutex
}

// NewSortedSet creates a new SortedSet instance that sorts its elements by the given weightVariable. It is possible to
// optionally provide a tieBreaker function that is used to break ties between elements with the same weight.
func newSortedSet[ElementType comparable, WeightType cmp.Ordered](weightVariable func(element ElementType) Variable[WeightType], optTieBreaker ...func(left, right ElementType) int) *sortedSet[ElementType, WeightType] {
	s := &sortedSet[ElementType, WeightType]{
		Set:             NewSet[ElementType](),
		elements:        shrinkingmap.New[ElementType, *sortedSetElement[ElementType, WeightType]](),
		sortedElements:  make([]*sortedSetElement[ElementType, WeightType], 0),
		heaviestElement: NewVariable[ElementType](),
		lightestElement: NewVariable[ElementType](),
		weightVariable:  weightVariable,
		tieBreaker:      lo.First(optTieBreaker, defaultTieBreaker[ElementType]),
	}

	s.OnUpdate(func(appliedMutations ds.SetMutations[ElementType]) {
		appliedMutations.AddedElements().Range(s.addSorted)
		appliedMutations.DeletedElements().Range(s.deleteSorted)
	})

	return s
}

// Ascending returns a slice of all elements of the set in ascending order.
func (s *sortedSet[ElementType, WeightType]) Ascending() (sortedSlice []ElementType) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if sortedElementsCount := len(s.sortedElements); sortedElementsCount > 0 {
		sortedSlice = make([]ElementType, sortedElementsCount)

		for i, sortedElement := range s.sortedElements {
			sortedSlice[sortedElementsCount-i-1] = sortedElement.element
		}
	}

	return sortedSlice
}

// Descending returns a slice of all elements of the set in descending order.
func (s *sortedSet[ElementType, WeightType]) Descending() (sortedSlice []ElementType) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if sortedElementsCount := len(s.sortedElements); sortedElementsCount > 0 {
		sortedSlice = make([]ElementType, sortedElementsCount)

		for i, sortedElement := range s.sortedElements {
			sortedSlice[i] = sortedElement.element
		}
	}

	return sortedSlice
}

// HeaviestElement returns the element with the heaviest weight.
func (s *sortedSet[ElementType, WeightType]) HeaviestElement() ReadableVariable[ElementType] {
	return s.heaviestElement
}

// LightestElement returns the element with the lightest weight.
func (s *sortedSet[ElementType, WeightType]) LightestElement() ReadableVariable[ElementType] {
	return s.lightestElement
}

// addSorted adds the given element to the sortedElements slice.
func (s *sortedSet[ElementType, WeightType]) addSorted(element ElementType) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if listElement, created := s.elements.GetOrCreate(element, func() *sortedSetElement[ElementType, WeightType] {
		return newSortedSetElement(element, s)
	}); created {
		listElement.unsubscribeFromWeightUpdates = s.weightVariable(element).OnUpdate(func(_ WeightType, newWeight WeightType) {
			// only lock if this is not the initial update
			if listElement.unsubscribeFromWeightUpdates != nil {
				s.mutex.Lock()
				defer s.mutex.Unlock()
			}

			listElement.weight = newWeight

			s.updatePosition(listElement)
		}, true)
	}
}

// deleteSorted deletes the given element from the sortedElements slice.
func (s *sortedSet[ElementType, WeightType]) deleteSorted(element ElementType) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if deletedElement, deleted := s.elements.DeleteAndReturn(element); deleted {
		// unsubscribe from weight updates
		deletedElement.unsubscribeFromWeightUpdates()

		// shift all elements to the right of the deleted element one position to the left
		for i := deletedElement.index; i < len(s.sortedElements)-1; i++ {
			s.sortedElements[i] = s.sortedElements[i+1]
			s.sortedElements[i].index--
		}

		// prevent memory leak and shrink slice
		s.sortedElements[len(s.sortedElements)-1] = nil
		s.sortedElements = s.sortedElements[:len(s.sortedElements)-1]

		// update heaviest and lightest element
		if deletedElement.index == 0 {
			if len(s.sortedElements) > 0 {
				s.heaviestElement.Set(s.sortedElements[0].element)
			} else {
				s.heaviestElement.Set(*new(ElementType))
			}
		}
		if deletedElement.index == len(s.sortedElements) {
			if len(s.sortedElements) > 0 {
				s.lightestElement.Set(s.sortedElements[len(s.sortedElements)-1].element)
			} else {
				s.lightestElement.Set(*new(ElementType))
			}
		}
	}
}

// updatePosition updates the position of the given element in the sortedElements slice.
func (s *sortedSet[ElementType, WeightType]) updatePosition(element *sortedSetElement[ElementType, WeightType]) (moved bool) {
	// update heaviest and lightest references after we are done moving the element
	defer func(oldIndex int) {
		if moved {
			if oldIndex == 0 {
				s.heaviestElement.Set(s.sortedElements[0].element)
			}

			if oldIndex == len(s.sortedElements)-1 {
				s.lightestElement.Set(s.sortedElements[len(s.sortedElements)-1].element)
			}
		}

		if element.index == 0 {
			s.heaviestElement.Set(element.element)
		}

		if element.index == len(s.sortedElements)-1 {
			s.lightestElement.Set(element.element)
		}
	}(element.index)

	// try to move the element to the left first
	for ; element.index != 0; moved = true {
		if !s.swap(s.sortedElements[element.index-1], element) {
			break
		}
	}

	// if the element was not moved to the left, try to move it to the right
	if !moved {
		for ; element.index != len(s.sortedElements)-1; moved = true {
			if !s.swap(element, s.sortedElements[element.index+1]) {
				break
			}
		}
	}

	return
}

// swap swaps the position of the two given elements in the sortedElements slice.
func (s *sortedSet[ElementType, WeightType]) swap(left *sortedSetElement[ElementType, WeightType], right *sortedSetElement[ElementType, WeightType]) (swapped bool) {
	if swapped = (left.weight < right.weight) || (left.weight == right.weight && s.tieBreaker(left.element, right.element) < 0); swapped {
		s.sortedElements[left.index], s.sortedElements[right.index] = s.sortedElements[right.index], s.sortedElements[left.index]
		left.index, right.index = right.index, left.index
	}

	return swapped
}

// defaultTieBreaker is the default tie-breaker function that is used to break ties between elements with the same weight.
func defaultTieBreaker[ElementType comparable](_ ElementType, _ ElementType) int {
	return 0
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region sortedSetElement /////////////////////////////////////////////////////////////////////////////////////////////

// sortedSetElement is an element of the sortedElements slice.
type sortedSetElement[ElementType comparable, WeightType cmp.Ordered] struct {
	// element is the element that is part of the set.
	element ElementType

	// weight is the weight of the element.
	weight WeightType

	// index is the index of the element in the sortedElements slice.
	index int

	// unsubscribeFromWeightUpdates is the function that is used to unsubscribe from weight updates.
	unsubscribeFromWeightUpdates func()
}

// newSortedSetElement creates a new sortedSetElement instance.
func newSortedSetElement[WeightType cmp.Ordered, ElementType comparable](element ElementType, sortedSet *sortedSet[ElementType, WeightType]) *sortedSetElement[ElementType, WeightType] {
	s := &sortedSetElement[ElementType, WeightType]{
		element: element,
		index:   len(sortedSet.sortedElements),
	}

	sortedSet.sortedElements = append(sortedSet.sortedElements, s)

	return s
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

package reactive

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SortedSet(t *testing.T) {
	element1 := newSortableElement("1st", 1)
	element2 := newSortableElement("2nd", 2)
	element3 := newSortableElement("3rd", 3)

	testSet := NewSortedSet((*sortableElement).weight)
	requireOrder(t, []*sortableElement{}, testSet)

	testSet.Add(element1)
	requireOrder(t, []*sortableElement{element1}, testSet)

	testSet.Add(element2)
	requireOrder(t, []*sortableElement{element2, element1}, testSet)

	testSet.Add(element3)
	requireOrder(t, []*sortableElement{element3, element2, element1}, testSet)

	element2.Weight.Set(5)
	requireOrder(t, []*sortableElement{element2, element3, element1}, testSet)

	element1.Weight.Set(4)
	requireOrder(t, []*sortableElement{element2, element1, element3}, testSet)

	element2.Weight.Set(3)
	requireOrder(t, []*sortableElement{element1, element3, element2}, testSet)

	testSet.Delete(element2)
	requireOrder(t, []*sortableElement{element1, element3}, testSet)

	testSet.Delete(element1)
	requireOrder(t, []*sortableElement{element3}, testSet)

	testSet.Add(element1)
	requireOrder(t, []*sortableElement{element1, element3}, testSet)

	testSet.Delete(element3)
	requireOrder(t, []*sortableElement{element1}, testSet)

	testSet.Delete(element1)
	requireOrder(t, []*sortableElement{}, testSet)
}

func requireOrder[ElementType comparable](t *testing.T, expectedElements []ElementType, sortedSet SortedSet[ElementType]) {
	descendingElements := sortedSet.Descending()
	require.Equal(t, len(expectedElements), len(descendingElements))
	for i, expectedElement := range expectedElements {
		require.Equal(t, expectedElement, descendingElements[i])
	}

	ascendingElements := sortedSet.Ascending()
	require.Equal(t, len(expectedElements), len(ascendingElements))
	for i, expectedElement := range expectedElements {
		require.Equal(t, expectedElement, ascendingElements[len(expectedElements)-i-1])
	}

	if len(expectedElements) > 0 {
		require.Equal(t, expectedElements[0], sortedSet.HeaviestElement().Get())
		require.Equal(t, expectedElements[len(expectedElements)-1], sortedSet.LightestElement().Get())
	} else {
		require.Equal(t, *new(ElementType), sortedSet.HeaviestElement().Get())
		require.Equal(t, *new(ElementType), sortedSet.LightestElement().Get())
	}
}

type sortableElement struct {
	value  string
	Weight Variable[int]
}

func newSortableElement(value string, weight int) *sortableElement {
	return &sortableElement{
		value:  value,
		Weight: NewVariable[int]().Init(weight),
	}
}

func (t *sortableElement) Less(other *sortableElement) bool {
	return t.value < other.value
}

func (t *sortableElement) weight() Variable[int] {
	return t.Weight
}

func (t *sortableElement) compare(other *sortableElement) int {
	switch {
	case t.value < other.value:
		return -1
	case t.value > other.value:
		return 1
	default:
		return 0
	}
}

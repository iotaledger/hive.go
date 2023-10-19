package ds

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	testList := NewList[int]()
	element1 := testList.PushBack(1)
	testList.PushBack(2)
	testList.PushBack(3)

	requireListElements(t, testList, []int{1, 2, 3})

	testList.MoveToBack(element1)
	requireListElements(t, testList, []int{2, 3, 1})

	testList.Remove(element1)
	requireListElements(t, testList, []int{2, 3})

	testList.PushBack(4)
	requireListElements(t, testList, []int{2, 3, 4})
}

func requireListElements[T any](t *testing.T, testList List[T], expectedValues []T) {
	require.Equal(t, len(expectedValues), testList.Len())

	testList.Range(func(value T) {
		expectedValue := expectedValues[0]
		expectedValues = expectedValues[1:]

		require.Equal(t, expectedValue, value)
	})
}

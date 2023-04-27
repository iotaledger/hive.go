package lo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Map(t *testing.T) {
	sourceSlice := []int{1, 2, 3}

	targetSlice := Map(sourceSlice, func(item int) int {
		return item * 2
	})

	require.Equal(t, []int{2, 4, 6}, targetSlice, "should map the slice")
}

func Test_Flatten(t *testing.T) {
	slices := [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}

	result := Flatten(slices)

	require.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, result, "should flatten the slices")
}

func Test_Reduce(t *testing.T) {
	collection := []int{1, 2, 3}

	result := Reduce(collection, func(accumulated int, item int) int {
		return accumulated + item
	}, 0)

	require.Equal(t, 6, result, "should reduce the slice")

}

func Test_Filter(t *testing.T) {
	collection := []int{1, 2, 3}

	result := Filter(collection, func(item int) bool {
		return item%2 == 0
	})

	require.Equal(t, []int{2}, result, "should filter the slice")
}

func Test_KeyBy(t *testing.T) {
	collection := []int{10, 20, 30}

	result := KeyBy(collection, func(item int) int {
		return item / 10
	})

	require.Equal(t, map[int]int{1: 10, 2: 20, 3: 30}, result, "should key the slice")
}

func Test_FilterByValue(t *testing.T) {
	collection := map[int]int{1: 10, 2: 20, 3: 30}

	result := FilterByValue(collection, func(item int) bool {
		return item%20 == 0
	})

	require.Equal(t, map[int]int{2: 20}, result, "should filter the slice")
}

func Test_Keys(t *testing.T) {
	collection := map[int]int{1: 10, 2: 20, 3: 30}

	result := Keys(collection)

	require.Contains(t, result, 1, "should get the keys")
	require.Contains(t, result, 2, "should get the keys")
	require.Contains(t, result, 3, "should get the keys")
}

func Test_Values(t *testing.T) {
	collection := map[int]int{1: 10, 2: 20, 3: 30}

	result := Values(collection)

	require.Contains(t, result, 10, "should get the values")
	require.Contains(t, result, 20, "should get the values")
	require.Contains(t, result, 30, "should get the values")
}

func Test_ForEach(t *testing.T) {
	collection := []int{1, 2, 3}

	result := []int{}

	ForEach(collection, func(item int) {
		result = append(result, item)
	})

	require.Equal(t, []int{1, 2, 3}, result, "should iterate over the slice")
}

func Test_ReduceProperty(t *testing.T) {
	collection := [][]int{{1, 2, 3}, {10, 20, 30}, {100, 200, 300}}

	result := ReduceProperty(collection, func(item []int) int {
		return item[0]
	}, func(accumulated int, item int) int {
		return accumulated + item
	}, 0)

	require.Equal(t, 111, result, "should reduce the slice")
}

func Test_Bind(t *testing.T) {
	f := func(param1, param2 int) int {
		return param1 * param2
	}

	boundF := Bind(10, f)

	require.Equal(t, 200, boundF(20), "should correcly add 10")
}

func Test_PanicOnErr(t *testing.T) {
	fPanic := func() (int, error) {
		return 0, fmt.Errorf("error")
	}

	fPanicCall := func() {
		PanicOnErr(fPanic())
	}
	require.Panics(t, fPanicCall, "should panic on error")

	fNoPanic := func() (int, error) {
		return 1, nil
	}

	fNoPanicCall := func() {
		val := PanicOnErr(fNoPanic())
		require.Equal(t, 1, val, "should return correct value")
	}
	require.NotPanics(t, fNoPanicCall, "should not panic without error")
}

func Test_Max(t *testing.T) {
	maxValueInt := Max(10, 1, 154, 61, 51, 65, 16, 51, 6, 516, 1, 65, -465, -465, -1, 0)

	require.Equal(t, 516, maxValueInt, "should correctly select maximum value")

	maxValueFloat := Max(1.0, -1.6, -1.5, -1.4, -1.3, -1.2, -1.1, -1.0, 1.5, 1.4, 1.3, 1.2, 1.1, 1.0)

	require.Equal(t, 1.5, maxValueFloat, "should correctly select maximum value")

	defaultIntValue := Max([]int{}...)

	require.Equal(t, 0, defaultIntValue, "should return default int value")
}

func Test_Min(t *testing.T) {
	maxValueInt := Min(10, 1, 154, 61, 51, 65, 16, 51, 6, 516, 1, 65, -465, -465, -1, 0)

	require.Equal(t, -465, maxValueInt, "should correctly select minimum value")

	maxValueFloat := Min(1.0, -1.6, -1.5, -1.4, -1.3, -1.2, -1.1, -1.0, 1.5, 1.4, 1.3, 1.2, 1.1, 1.0)

	require.Equal(t, -1.6, maxValueFloat, "should correctly select minimum value")

	defaultIntValue := Min([]int{}...)

	require.Equal(t, 0, defaultIntValue, "should return default int value")
}

func Test_Sum(t *testing.T) {
	maxValueInt := Sum(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	require.Equal(t, 55, maxValueInt, "should correctly sum values")
}

package ringbuffer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRingBuffer(t *testing.T) {
	r := NewRingBuffer[int](3)

	require.Equal(t, []int{}, r.ToSlice())

	r.Add(1)
	r.Add(2)
	expected := []int{2, 1}
	actual := r.ToSlice()
	require.Len(t, actual, 2)
	require.Equal(t, expected, actual, "Buffer should contain %v but got %v", expected, actual)

	r.Add(3)

	// Test that the buffer contains the expected elements
	expected = []int{3, 2, 1}
	actual = r.ToSlice()
	require.Equal(t, expected, actual, "Buffer should contain %v but got %v", expected, actual)

	// Add 2 more elements to the buffer, overwriting the oldest element
	r.Add(4)
	r.Add(5)

	// Test that the buffer contains the expected elements
	expected = []int{5, 4, 3}
	actual = r.ToSlice()
	require.Equal(t, expected, actual, "Buffer should contain %v but got %v", expected, actual)

	// Add 1 more element to the buffer, overwriting the oldest element
	r.Add(6)

	// Test that the buffer contains the expected elements
	expected = []int{6, 5, 4}
	actual = r.ToSlice()
	require.Equal(t, expected, actual, "Buffer should contain %v but got %v", expected, actual)
}

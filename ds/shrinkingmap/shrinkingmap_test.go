package shrinkingmap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShrinkingMap_Ratio(t *testing.T) {
	shrink := New[int, int](
		WithShrinkingThresholdRatio(2.0),
		WithShrinkingThresholdCount(0),
	)
	for i := 0; i < 100; i++ {
		shrink.Set(i, i)
	}

	require.Equal(t, 0, shrink.deletedKeys)

	for i := 0; i < 67; i++ {
		require.Equal(t, i, shrink.deletedKeys)
		shrink.Delete(i)
	}

	require.Equal(t, 0, shrink.deletedKeys)
	shrink.Delete(99)
	require.Equal(t, 1, shrink.deletedKeys)
	shrink.Shrink()
	require.Equal(t, 0, shrink.deletedKeys)
}

func TestShrinkingMap_Count(t *testing.T) {
	shrink := New[int, int](
		WithShrinkingThresholdRatio(0.0),
		WithShrinkingThresholdCount(10),
	)
	for i := 0; i < 100; i++ {
		shrink.Set(i, i)
	}

	require.Equal(t, 0, shrink.deletedKeys)

	for i := 0; i < 10; i++ {
		require.Equal(t, i, shrink.deletedKeys)
		shrink.Delete(i)
	}

	require.Equal(t, 0, shrink.deletedKeys)
	shrink.Delete(99)
	require.Equal(t, 1, shrink.deletedKeys)
	shrink.Shrink()
	require.Equal(t, 0, shrink.deletedKeys)
}

func TestShrinkingMap_Both(t *testing.T) {
	// check count condition reached after ratio
	shrink := New[int, int](
		WithShrinkingThresholdRatio(2.0),
		WithShrinkingThresholdCount(70),
	)
	for i := 0; i < 100; i++ {
		shrink.Set(i, i)
	}

	require.Equal(t, 0, shrink.deletedKeys)

	for i := 0; i < 70; i++ {
		require.Equal(t, i, shrink.deletedKeys)
		shrink.Delete(i)
	}

	require.Equal(t, 0, shrink.deletedKeys)
	shrink.Delete(99)
	require.Equal(t, 1, shrink.deletedKeys)
	shrink.Shrink()
	require.Equal(t, 0, shrink.deletedKeys)

	// check ratio condition reached after count
	shrink = New[int, int](
		WithShrinkingThresholdRatio(2.0),
		WithShrinkingThresholdCount(60),
	)
	for i := 0; i < 100; i++ {
		shrink.Set(i, i)
	}

	require.Equal(t, 0, shrink.deletedKeys)

	for i := 0; i < 67; i++ {
		require.Equal(t, i, shrink.deletedKeys)
		shrink.Delete(i)
	}

	require.Equal(t, 0, shrink.deletedKeys)
	shrink.Delete(99)
	require.Equal(t, 1, shrink.deletedKeys)
	shrink.Shrink()
	require.Equal(t, 0, shrink.deletedKeys)
}

func TestShrinkingMap_Empty(t *testing.T) {

	// check count condition reached after ratio
	shrink := New[int, int](
		WithShrinkingThresholdRatio(2.0),
		WithShrinkingThresholdCount(70),
	)

	require.True(t, shrink.IsEmpty())

	for i := 0; i < 100; i++ {
		shrink.Set(i, i)
	}
	require.Equal(t, 100, shrink.Size())
	require.False(t, shrink.IsEmpty())

	for i := 0; i < 100; i++ {
		shrink.Delete(i)
	}

	require.Equal(t, 0, shrink.Size())
	require.True(t, shrink.IsEmpty())
	require.True(t, shrink.deletedKeys > 0)
}

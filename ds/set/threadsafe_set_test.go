//nolint:unparam // we don't care about these linters in test cases
package set

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/izuc/zipp.foundation/serializer/v2/serix"
)

func TestThreadSafeSet_Add(t *testing.T) {
	set := initThreadSafeSet(3, 0)

	require.Equal(t, 3, set.Size(), "wrong size")
	require.False(t, set.Add("item1"), "the item should already exist")
	require.Equal(t, 3, set.Size(), "wrong size")
	require.True(t, set.Add("item4"), "the item should not exist")
	require.Equal(t, 4, set.Size(), "wrong size")
}

func TestThreadSafeSet_Delete(t *testing.T) {
	set := initThreadSafeSet(3, 0)

	require.Equal(t, 3, set.Size(), "wrong size")
	require.True(t, set.Delete("item2"), "the element should exist")
	require.Equal(t, 2, set.Size(), "wrong size")
	require.False(t, set.Delete("item2"), "the element should not exist")
	require.Equal(t, 2, set.Size(), "wrong size")
}

func TestThreadSafeSet_Has(t *testing.T) {
	set := initThreadSafeSet(3, 0)

	require.True(t, set.Has("item2"), "the element should exist")
	require.True(t, set.Delete("item2"), "the element should exist")
	require.False(t, set.Has("item2"), "the element should not exist")
	require.True(t, set.Delete("item1"), "the element should exist")
	require.False(t, set.Has("item1"), "the element should not exist")
}

func TestThreadSafeSet_ForEach(t *testing.T) {
	set := initThreadSafeSet(3, 0)

	expectedElements := initThreadSafeSet(3, 0)
	require.Equal(t, 3, expectedElements.Size(), "wrong size")
	set.ForEach(func(element string) {
		require.True(t, expectedElements.Delete(element))
	})
	require.Equal(t, 0, expectedElements.Size(), "wrong size")
}

func TestThreadSafeSet_Clear(t *testing.T) {
	set := initThreadSafeSet(3, 0)
	require.Equal(t, 3, set.Size(), "wrong size")

	set.Clear()
	require.Equal(t, 0, set.Size(), "wrong size")
}

func TestThreadSafeSet_Size(t *testing.T) {
	set := initThreadSafeSet(3, 0)
	require.Equal(t, 3, set.Size(), "wrong size")
	set = initSimpleSet(0, 0)
	require.Equal(t, 0, set.Size(), "wrong size")
	set = initSimpleSet(100000, 0)
	require.Equal(t, 100000, set.Size(), "wrong size")
}

func TestThreadSafeSet_Encoding(t *testing.T) {
	serix.DefaultAPI.RegisterTypeSettings("", serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte))

	set := initThreadSafeSet(3, 0)
	bytes, err := set.Encode()
	require.NoError(t, err)

	decoded := new(threadSafeSet[string])
	consumed, err := decoded.Decode(bytes)
	require.NoError(t, err)
	require.Equal(t, len(bytes), consumed)

	require.Equal(t, set, decoded)
}

func initThreadSafeSet(count int, start int) Set[string] {
	set := newThreadSafeSet[string]()
	end := start + count
	for i := start; i < end; i++ {
		set.Add(fmt.Sprintf("item%d", i))
	}

	return set
}

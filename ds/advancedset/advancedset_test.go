package advancedset

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/izuc/zipp.foundation/serializer/v2/serix"
)

func TestAdvancedSet_IsEmpty(t *testing.T) {
	set := initAdvancedSet(1, 0)

	require.False(t, set.IsEmpty(), "the set should not be empty")
	require.True(t, set.Delete("item0"), "the item should already exist")
	require.True(t, true, set.IsEmpty(), "the set should be empty")
}

func TestAdvancedSet_Add(t *testing.T) {
	set := initAdvancedSet(3, 0)

	require.Equal(t, 3, set.Size(), "wrong size")
	require.False(t, set.Add("item1"), "the item should already exist")
	require.Equal(t, 3, set.Size(), "wrong size")
	require.True(t, set.Add("item4"), "the item should not exist")
	require.Equal(t, 4, set.Size(), "wrong size")
}

func TestAdvancedSet_AddAll(t *testing.T) {
	set := initAdvancedSet(3, 0)
	set2 := initAdvancedSet(3, 4)

	require.Equal(t, 3, set.Size(), "wrong size")
	require.True(t, set.AddAll(set2), "should add elements to the set")
	require.Equal(t, 6, set.Size(), "wrong size")
}

func TestAdvancedSet_DeleteAll(t *testing.T) {
	set := initAdvancedSet(3, 0)
	set2 := initAdvancedSet(3, 1)

	require.Equal(t, 3, set.Size(), "wrong size")
	require.Equal(t, 2, set.DeleteAll(set2).Size(), "should remove 2 elements from the set")
	require.Equal(t, 1, set.Size(), "wrong size")
}

func TestAdvancedSet_Delete(t *testing.T) {
	set := initAdvancedSet(3, 0)

	require.Equal(t, 3, set.Size(), "wrong size")
	require.True(t, set.Delete("item2"), "the element should exist")
	require.Equal(t, 2, set.Size(), "wrong size")
	require.False(t, set.Delete("item2"), "the element should not exist")
	require.Equal(t, 2, set.Size(), "wrong size")
}

func TestAdvancedSet_Has(t *testing.T) {
	set := initAdvancedSet(3, 0)

	require.True(t, set.Has("item2"), "the element should exist")
	require.True(t, set.Delete("item2"), "the element should exist")
	require.False(t, set.Has("item2"), "the element should not exist")
	require.True(t, set.Delete("item1"), "the element should exist")
	require.False(t, set.Has("item1"), "the element should not exist")
}

func TestAdvancedSet_HasAll(t *testing.T) {
	set := initAdvancedSet(3, 0)
	set2 := initAdvancedSet(2, 1)

	require.True(t, set.HasAll(set2), "all elements should exist")
	require.True(t, set.Delete("item2"), "the element should exist")
	require.False(t, set.HasAll(set2), "the elements should not exist")
	require.True(t, set.Delete("item1"), "the element should exist")
	require.False(t, set.HasAll(set2), "the elements should not exist")
}

func TestAdvancedSet_ForEach(t *testing.T) {
	set := initAdvancedSet(3, 0)

	expectedElements := initAdvancedSet(3, 0)
	require.Equal(t, 3, expectedElements.Size(), "wrong size")
	err := set.ForEach(func(element string) error {
		require.True(t, expectedElements.Delete(element))

		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 0, expectedElements.Size(), "wrong size")
}

func TestAdvancedSet_Intersect(t *testing.T) {
	set := initAdvancedSet(5, 0)
	set2 := initAdvancedSet(5, 3)

	intersectExpected := initAdvancedSet(2, 3)

	require.True(t, set.Intersect(set2).HasAll(intersectExpected), "wrong intersection")
}

func TestAdvancedSet_Filter(t *testing.T) {
	set := initAdvancedSet(5, 0)

	require.True(t, set.Filter(func(elem string) bool { return elem[4:] == "3" }).Is("item3"), "wrong filter result")
}

func TestAdvancedSet_Equal(t *testing.T) {
	set := initAdvancedSet(3, 0)
	set2 := initAdvancedSet(3, 0)

	require.True(t, set.Equal(set2), "the sets should be equal")

}

func TestAdvancedSet_Clone(t *testing.T) {
	set := initAdvancedSet(3, 0)
	set2 := set.Clone()

	require.True(t, set.Add("item6"), "the item should not exist")
	require.True(t, set2.Delete("item0"), "the item should not exist")
	require.False(t, set.Equal(set2), "the sets should be equal")
}

func TestAdvancedSet_Slice(t *testing.T) {
	set := initAdvancedSet(3, 0)
	setSlice := set.Slice()

	require.Equal(t, set.Size(), len(setSlice), "length should be equal")
	require.True(t, New(setSlice...).Equal(set), "sets should be equal")
}

func TestAdvancedSet_Iterator(t *testing.T) {
	set := initAdvancedSet(3, 0)
	setWalker := set.Iterator()
	counter := 0
	for setWalker.HasNext() {
		require.True(t, set.Has(setWalker.Next()), "the element should exist in original set")
		counter++
	}

	require.Equal(t, set.Size(), counter, "should walk through all elements")
}

func TestAdvancedSet_Clear(t *testing.T) {
	set := initAdvancedSet(3, 0)
	require.Equal(t, 3, set.Size(), "wrong size")

	set.Clear()
	require.Equal(t, 0, set.Size(), "wrong size")
}

func TestAdvancedSet_Size(t *testing.T) {
	set := initAdvancedSet(3, 0)
	require.Equal(t, 3, set.Size(), "wrong size")
	set = initAdvancedSet(0, 0)
	require.Equal(t, 0, set.Size(), "wrong size")
	set = initAdvancedSet(100000, 0)
	require.Equal(t, 100000, set.Size(), "wrong size")
}

func TestAdvancedSet_Encoding(t *testing.T) {
	serix.DefaultAPI.RegisterTypeSettings("", serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte))

	set := initAdvancedSet(3, 0)
	bytes, err := set.Encode()
	require.NoError(t, err)

	decoded := new(AdvancedSet[string])
	consumed, err := decoded.Decode(bytes)
	require.NoError(t, err)
	require.Equal(t, len(bytes), consumed)

	require.Equal(t, set, decoded)
}

func initAdvancedSet(count int, start int) *AdvancedSet[string] {
	set := New[string]()
	end := start + count
	for i := start; i < end; i++ {
		set.Add(fmt.Sprintf("item%d", i))
	}

	return set
}

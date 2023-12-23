package ds_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/ds"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

func TestSet_IsEmpty(t *testing.T) {
	set := initSet(1, 0)

	require.False(t, set.IsEmpty(), "the set should not be empty")
	require.True(t, set.Delete("item0"), "the item should already exist")
	require.True(t, true, set.IsEmpty(), "the set should be empty")
}

func TestSet_Add(t *testing.T) {
	set := initSet(3, 0)

	require.Equal(t, 3, set.Size(), "wrong size")
	require.False(t, set.Add("item1"), "the item should already exist")
	require.Equal(t, 3, set.Size(), "wrong size")
	require.True(t, set.Add("item4"), "the item should not exist")
	require.Equal(t, 4, set.Size(), "wrong size")
}

func TestSet_AddAll(t *testing.T) {
	set := initSet(3, 0)
	set2 := initSet(3, 4)

	require.Equal(t, 3, set.Size(), "wrong size")
	require.True(t, set.AddAll(set2).HasAll(set2), "should add elements to the set")
	require.Equal(t, 6, set.Size(), "wrong size")
}

func TestSet_DeleteAll(t *testing.T) {
	set := initSet(3, 0)
	set2 := initSet(3, 1)

	require.Equal(t, 3, set.Size(), "wrong size")
	require.Equal(t, 2, set.DeleteAll(set2).Size(), "should remove 2 elements from the set")
	require.Equal(t, 1, set.Size(), "wrong size")
}

func TestSet_Delete(t *testing.T) {
	set := initSet(3, 0)

	require.Equal(t, 3, set.Size(), "wrong size")
	require.True(t, set.Delete("item2"), "the element should exist")
	require.Equal(t, 2, set.Size(), "wrong size")
	require.False(t, set.Delete("item2"), "the element should not exist")
	require.Equal(t, 2, set.Size(), "wrong size")
}

func TestSet_Has(t *testing.T) {
	set := initSet(3, 0)

	require.True(t, set.Has("item2"), "the element should exist")
	require.True(t, set.Delete("item2"), "the element should exist")
	require.False(t, set.Has("item2"), "the element should not exist")
	require.True(t, set.Delete("item1"), "the element should exist")
	require.False(t, set.Has("item1"), "the element should not exist")
}

func TestSet_HasAll(t *testing.T) {
	set := initSet(3, 0)
	set2 := initSet(2, 1)

	require.True(t, set.HasAll(set2), "all elements should exist")
	require.True(t, set.Delete("item2"), "the element should exist")
	require.False(t, set.HasAll(set2), "the elements should not exist")
	require.True(t, set.Delete("item1"), "the element should exist")
	require.False(t, set.HasAll(set2), "the elements should not exist")
}

func TestSet_ForEach(t *testing.T) {
	set := initSet(3, 0)

	expectedElements := initSet(3, 0)
	require.Equal(t, 3, expectedElements.Size(), "wrong size")
	err := set.ForEach(func(element string) error {
		require.True(t, expectedElements.Delete(element))

		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 0, expectedElements.Size(), "wrong size")
}

func TestSet_RangeAndString(t *testing.T) {
	set := initSet(3, 0)

	expectedElements := initSet(3, 0)
	require.Equal(t, 3, expectedElements.Size(), "wrong size")

	str := set.String()
	set.Range(func(element string) {
		require.Contains(t, str, element)
	})
}

func TestSet_Intersect(t *testing.T) {
	set := initSet(5, 0)
	set2 := initSet(5, 3)

	intersectExpected := initSet(2, 3)

	require.True(t, set.Intersect(set2).HasAll(intersectExpected), "wrong intersection")
}

func TestSet_Filter(t *testing.T) {
	set := initSet(5, 0)

	require.True(t, set.Filter(func(elem string) bool { return elem[4:] == "3" }).Is("item3"), "wrong filter result")
}

func TestSet_Equal(t *testing.T) {
	set := initSet(3, 0)
	set2 := initSet(3, 0)

	require.True(t, set.Equals(set2), "the sets should be equal")

}

func TestSet_Clone(t *testing.T) {
	set := initSet(3, 0)
	set2 := set.Clone()

	require.True(t, set.Add("item6"), "the item should not exist")
	require.True(t, set2.Delete("item0"), "the item should not exist")
	require.False(t, set.Equals(set2), "the sets should be equal")
}

func TestSet_Slice(t *testing.T) {
	testSet := initSet(3, 0)
	setSlice := testSet.ToSlice()

	require.Equal(t, testSet.Size(), len(setSlice), "length should be equal")
	require.True(t, ds.NewSet(setSlice...).Equals(testSet), "sets should be equal")
}

func TestSet_Iterator(t *testing.T) {
	set := initSet(3, 0)
	setWalker := set.Iterator()
	counter := 0
	for setWalker.HasNext() {
		require.True(t, set.Has(setWalker.Next()), "the element should exist in original set")
		counter++
	}

	require.Equal(t, set.Size(), counter, "should walk through all elements")
}

func TestSet_Clear(t *testing.T) {
	set := initSet(3, 0)
	require.Equal(t, 3, set.Size(), "wrong size")

	set.Clear()
	require.Equal(t, 0, set.Size(), "wrong size")
}

func TestSet_Size(t *testing.T) {
	set := initSet(3, 0)
	require.Equal(t, 3, set.Size(), "wrong size")
	set = initSet(0, 0)
	require.Equal(t, 0, set.Size(), "wrong size")
	set = initSet(100000, 0)
	require.Equal(t, 100000, set.Size(), "wrong size")
}

func TestSet_Encoding(t *testing.T) {
	serix.DefaultAPI.RegisterTypeSettings("", serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte))

	testSet := initSet(3, 0)
	bytes, err := testSet.Encode(serix.DefaultAPI)
	require.NoError(t, err)

	decoded := ds.NewSet[string]()
	consumed, err := decoded.Decode(serix.DefaultAPI, bytes)
	require.NoError(t, err)
	require.Equal(t, len(bytes), consumed)

	require.Equal(t, testSet, decoded)
}

func initSet(count int, start int) ds.Set[string] {
	set := ds.NewSet[string]()
	end := start + count
	for i := start; i < end; i++ {
		set.Add(fmt.Sprintf("item%d", i))
	}

	return set
}

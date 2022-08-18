package set

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdvancedSet_IsEmpty(t *testing.T) {
	set := initAdvancedSet(1, 0)

	assert.False(t, set.IsEmpty(), "the set should not be empty")
	assert.True(t, set.Delete("item0"), "the item should already exist")
	assert.True(t, true, set.IsEmpty(), "the set should be empty")
}

func TestAdvancedSet_Add(t *testing.T) {
	set := initAdvancedSet(3, 0)

	assert.Equal(t, 3, set.Size(), "wrong size")
	assert.False(t, set.Add("item1"), "the item should already exist")
	assert.Equal(t, 3, set.Size(), "wrong size")
	assert.True(t, set.Add("item4"), "the item should not exist")
	assert.Equal(t, 4, set.Size(), "wrong size")
}

func TestAdvancedSet_AddAll(t *testing.T) {
	set := initAdvancedSet(3, 0)
	set2 := initAdvancedSet(3, 4)

	assert.Equal(t, 3, set.Size(), "wrong size")
	assert.True(t, set.AddAll(set2), "should add elements to the set")
	assert.Equal(t, 6, set.Size(), "wrong size")
}

func TestAdvancedSet_DeleteAll(t *testing.T) {
	set := initAdvancedSet(3, 0)
	set2 := initAdvancedSet(3, 1)

	assert.Equal(t, 3, set.Size(), "wrong size")
	assert.Equal(t, 2, set.DeleteAll(set2).Size(), "should remove 2 elements from the set")
	assert.Equal(t, 1, set.Size(), "wrong size")
}

func TestAdvancedSet_Delete(t *testing.T) {
	set := initAdvancedSet(3, 0)

	assert.Equal(t, 3, set.Size(), "wrong size")
	assert.True(t, set.Delete("item2"), "the element should exist")
	assert.Equal(t, 2, set.Size(), "wrong size")
	assert.False(t, set.Delete("item2"), "the element should not exist")
	assert.Equal(t, 2, set.Size(), "wrong size")
}

func TestAdvancedSet_Has(t *testing.T) {
	set := initAdvancedSet(3, 0)

	assert.True(t, set.Has("item2"), "the element should exist")
	assert.True(t, set.Delete("item2"), "the element should exist")
	assert.False(t, set.Has("item2"), "the element should not exist")
	assert.True(t, set.Delete("item1"), "the element should exist")
	assert.False(t, set.Has("item1"), "the element should not exist")
}

func TestAdvancedSet_HasAll(t *testing.T) {
	set := initAdvancedSet(3, 0)
	set2 := initAdvancedSet(2, 1)

	assert.True(t, set.HasAll(set2), "all elements should exist")
	assert.True(t, set.Delete("item2"), "the element should exist")
	assert.False(t, set.HasAll(set2), "the elements should not exist")
	assert.True(t, set.Delete("item1"), "the element should exist")
	assert.False(t, set.HasAll(set2), "the elements should not exist")
}

func TestAdvancedSet_ForEach(t *testing.T) {
	set := initAdvancedSet(3, 0)

	expectedElements := initAdvancedSet(3, 0)
	assert.Equal(t, 3, expectedElements.Size(), "wrong size")
	err := set.ForEach(func(element string) error {
		assert.True(t, expectedElements.Delete(element))

		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, expectedElements.Size(), "wrong size")
}

func TestAdvancedSet_Intersect(t *testing.T) {
	set := initAdvancedSet(5, 0)
	set2 := initAdvancedSet(5, 3)

	intersectExpected := initAdvancedSet(2, 3)

	assert.True(t, set.Intersect(set2).HasAll(intersectExpected), "wrong intersection")
}

func TestAdvancedSet_Filter(t *testing.T) {
	set := initAdvancedSet(5, 0)

	assert.True(t, set.Filter(func(elem string) bool { return elem[4:] == "3" }).Is("item3"), "wrong filter result")
}

func TestAdvancedSet_Equal(t *testing.T) {
	set := initAdvancedSet(3, 0)
	set2 := initAdvancedSet(3, 0)

	assert.True(t, set.Equal(set2), "the sets should be equal")

}

func TestAdvancedSet_Clone(t *testing.T) {
	set := initAdvancedSet(3, 0)
	set2 := set.Clone()

	assert.True(t, set.Add("item6"), "the item should not exist")
	assert.True(t, set2.Delete("item0"), "the item should not exist")
	assert.False(t, set.Equal(set2), "the sets should be equal")
}

func TestAdvancedSet_Slice(t *testing.T) {
	set := initAdvancedSet(3, 0)
	setSlice := set.Slice()

	assert.Equal(t, set.Size(), len(setSlice), "length should be equal")
	assert.True(t, NewAdvancedSet(setSlice...).Equal(set), "sets should be equal")
}

func TestAdvancedSet_Iterator(t *testing.T) {
	set := initAdvancedSet(3, 0)
	setWalker := set.Iterator()
	counter := 0
	for setWalker.HasNext() {
		assert.True(t, set.Has(setWalker.Next()), "the element should exist in original set")
		counter++
	}

	assert.Equal(t, set.Size(), counter, "should walk through all elements")
}

func TestAdvancedSet_Clear(t *testing.T) {
	set := initAdvancedSet(3, 0)
	assert.Equal(t, 3, set.Size(), "wrong size")

	set.Clear()
	assert.Equal(t, 0, set.Size(), "wrong size")
}

func TestAdvancedSet_Size(t *testing.T) {
	set := initAdvancedSet(3, 0)
	assert.Equal(t, 3, set.Size(), "wrong size")
	set = initAdvancedSet(0, 0)
	assert.Equal(t, 0, set.Size(), "wrong size")
	set = initAdvancedSet(100000, 0)
	assert.Equal(t, 100000, set.Size(), "wrong size")
}

func initAdvancedSet(count int, start int) *AdvancedSet[string] {
	set := NewAdvancedSet[string]()
	end := start + count
	for i := start; i < end; i++ {
		set.Add(fmt.Sprintf("item%d", i))
	}

	return set
}

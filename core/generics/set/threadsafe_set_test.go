package set

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/core/datastructure/set"
)

func TestThreadSafeSet_Add(t *testing.T) {
	set := initThreadSafeSet(3, 0)

	assert.Equal(t, 3, set.Size(), "wrong size")
	assert.False(t, set.Add("item1"), "the item should already exist")
	assert.Equal(t, 3, set.Size(), "wrong size")
	assert.True(t, set.Add("item4"), "the item should not exist")
	assert.Equal(t, 4, set.Size(), "wrong size")
}

func TestThreadSafeSet_Delete(t *testing.T) {
	set := initThreadSafeSet(3, 0)

	assert.Equal(t, 3, set.Size(), "wrong size")
	assert.True(t, set.Delete("item2"), "the element should exist")
	assert.Equal(t, 2, set.Size(), "wrong size")
	assert.False(t, set.Delete("item2"), "the element should not exist")
	assert.Equal(t, 2, set.Size(), "wrong size")
}

func TestThreadSafeSet_Has(t *testing.T) {
	set := initThreadSafeSet(3, 0)

	assert.True(t, set.Has("item2"), "the element should exist")
	assert.True(t, set.Delete("item2"), "the element should exist")
	assert.False(t, set.Has("item2"), "the element should not exist")
	assert.True(t, set.Delete("item1"), "the element should exist")
	assert.False(t, set.Has("item1"), "the element should not exist")
}

func TestThreadSafeSet_ForEach(t *testing.T) {
	set := initThreadSafeSet(3, 0)

	expectedElements := initThreadSafeSet(3, 0)
	assert.Equal(t, 3, expectedElements.Size(), "wrong size")
	set.ForEach(func(element string) {
		assert.True(t, expectedElements.Delete(element))
	})
	assert.Equal(t, 0, expectedElements.Size(), "wrong size")
}

func TestThreadSafeSet_Clear(t *testing.T) {
	set := initThreadSafeSet(3, 0)
	assert.Equal(t, 3, set.Size(), "wrong size")

	set.Clear()
	assert.Equal(t, 0, set.Size(), "wrong size")
}

func TestThreadSafeSet_Size(t *testing.T) {
	set := initThreadSafeSet(3, 0)
	assert.Equal(t, 3, set.Size(), "wrong size")
	set = initSimpleSet(0, 0)
	assert.Equal(t, 0, set.Size(), "wrong size")
	set = initSimpleSet(100000, 0)
	assert.Equal(t, 100000, set.Size(), "wrong size")
}

func initThreadSafeSet(count int, start int) Set[string] {
	set := newGenericSet[string](set.New(true))
	end := start + count
	for i := start; i < end; i++ {
		set.Add(fmt.Sprintf("item%d", i))
	}

	return set
}

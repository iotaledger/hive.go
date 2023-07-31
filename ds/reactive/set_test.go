package reactive

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/ds"
)

func TestSet(t *testing.T) {
	source1 := NewSet[int]()
	source2 := NewSet[int]()

	inheritedSet := NewDerivedSet[int]()
	inheritedSet.InheritFrom(source1, source2)

	source1.AddAll(ds.NewSet(1, 2, 4))
	source2.AddAll(ds.NewSet(7, 9))

	require.True(t, inheritedSet.Has(1))
	require.True(t, inheritedSet.Has(2))
	require.True(t, inheritedSet.Has(4))
	require.True(t, inheritedSet.Has(7))
	require.True(t, inheritedSet.Has(9))

	inheritedSet1 := NewDerivedSet[int]()
	inheritedSet1.InheritFrom(source1, source2)

	require.True(t, inheritedSet1.Has(1))
	require.True(t, inheritedSet1.Has(2))
	require.True(t, inheritedSet1.Has(4))
	require.True(t, inheritedSet1.Has(7))
	require.True(t, inheritedSet1.Has(9))
}

func TestSubtract(t *testing.T) {
	sourceSet := NewSet[int]()
	sourceSet.Add(3)

	removedSet := NewSet[int]()
	removedSet.Add(5)

	subtraction := sourceSet.SubtractReactive(removedSet)
	require.True(t, subtraction.Has(3))
	require.Equal(t, 1, subtraction.Size())

	sourceSet.Add(4)
	require.True(t, subtraction.Has(3))
	require.True(t, subtraction.Has(4))
	require.Equal(t, 2, subtraction.Size())

	removedSet.Add(4)
	require.True(t, subtraction.Has(3))
	require.False(t, subtraction.Has(4))
	require.Equal(t, 1, subtraction.Size())

	removedSet.Add(3)
	require.False(t, subtraction.Has(3))
	require.False(t, subtraction.Has(4))
	require.Equal(t, 0, subtraction.Size())

	sourceSet.Add(5)
	require.False(t, subtraction.Has(3))
	require.False(t, subtraction.Has(4))
	require.False(t, subtraction.Has(3))
	require.False(t, subtraction.Has(5))
	require.Equal(t, 0, subtraction.Size())

	sourceSet.Add(6)
	require.False(t, subtraction.Has(3))
	require.False(t, subtraction.Has(4))
	require.False(t, subtraction.Has(3))
	require.False(t, subtraction.Has(5))
	require.True(t, subtraction.Has(6))
	require.Equal(t, 1, subtraction.Size())
}

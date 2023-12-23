package reactive

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/ds"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

func TestSet(t *testing.T) {
	source1 := NewSet[int]()
	source2 := NewSet[int]()

	inheritedSet := NewDerivedSet[int]()
	defer inheritedSet.InheritFrom(source1, source2)()

	source1.AddAll(ds.NewSet(1, 2, 4))
	source2.AddAll(ds.NewSet(7, 9))

	require.True(t, inheritedSet.Has(1))
	require.True(t, inheritedSet.Has(2))
	require.True(t, inheritedSet.Has(4))
	require.True(t, inheritedSet.Has(7))
	require.True(t, inheritedSet.Has(9))

	inheritedSet1 := NewDerivedSet[int]()
	defer inheritedSet1.InheritFrom(source1, source2)()

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

func TestNewSet(t *testing.T) {
	s := newSet[int](1, 2, 3)
	require.NotNil(t, s)
	require.Equal(t, 3, s.Size())
}

func TestSet_Add(t *testing.T) {
	s := newSet[int]()
	added := s.Add(1)
	require.True(t, added)
	require.Equal(t, 1, s.Size())

	// Test adding a duplicate element
	added = s.Add(1)
	require.False(t, added)
	require.Equal(t, 1, s.Size())
}

func TestSet_AddAll(t *testing.T) {
	s := newSet[int]()
	otherSet := newSet[int](2, 3)
	addedElements := s.AddAll(otherSet.ReadOnly())

	require.Equal(t, 2, addedElements.Size())
	require.Equal(t, 2, s.Size())

	// Test adding duplicate elements
	addedElements = s.AddAll(otherSet.ReadOnly())
	require.Equal(t, 0, addedElements.Size())
	require.Equal(t, 2, s.Size())
}

func TestSet_Delete(t *testing.T) {
	s := newSet[int](1, 2, 3)
	deleted := s.Delete(2)
	require.True(t, deleted)
	require.Equal(t, 2, s.Size())

	// Test deleting a non-existent element
	deleted = s.Delete(4)
	require.False(t, deleted)
	require.Equal(t, 2, s.Size())
}

func TestSet_DeleteAll(t *testing.T) {
	s := newSet[int](1, 2, 3, 4)
	otherSet := newSet[int](2, 3)
	deletedElements := s.DeleteAll(otherSet.ReadOnly())

	require.Equal(t, 2, deletedElements.Size())
	require.Equal(t, 2, s.Size())
}

func TestNewReadableSet(t *testing.T) {
	r := newReadableSet[int](1, 2, 3)
	require.NotNil(t, r)
	require.Equal(t, 3, r.Size())
}

func TestNewDerivedSet(t *testing.T) {
	d := newDerivedSet[int]()
	require.NotNil(t, d)
}

// TestReadableSet_SubtractReactive tests the SubtractReactive method
func TestReadableSet_SubtractReactive(t *testing.T) {
	mainSet := newReadableSet[int](1, 2, 3)
	subtractSet := newSet[int](2)

	result := mainSet.SubtractReactive(subtractSet)

	// Initially, 2 should be subtracted from mainSet
	require.Equal(t, 3, mainSet.Size())
	require.False(t, result.Has(2))
	require.True(t, result.Has(1))
	require.True(t, result.Has(3))
	require.Equal(t, 1, subtractSet.Size())
	require.Equal(t, 2, result.Size())

	// Update subtractSet and check if result set is updated
	subtractSet.Add(3)
	require.False(t, result.Has(3)) // Now 3 should also be subtracted
	require.Equal(t, 1, result.Size())
	require.Equal(t, 3, mainSet.Size())
	require.Equal(t, 2, subtractSet.Size())
}

// TestReadableSet_WithElements tests the WithElements method
func TestReadableSet_WithElements(t *testing.T) {
	rSet := newSet[int](1, 2, 3)

	setupCalledTimes := 0
	teardownCalledTimes := 0

	teardown := rSet.WithElements(
		func(element int) (teardown func()) {
			setupCalledTimes++
			return func() {
				teardownCalledTimes++
			}
		},
	)
	defer teardown()

	rSet.Add(4) // Trigger setup
	rSet.Add(4) // Should not trigger setup again as it is a duplicate
	rSet.Add(1) // Should not trigger setup again as it is a duplicate
	require.Equal(t, 4, setupCalledTimes)

	rSet.Delete(4) // Trigger teardown
	require.Equal(t, 1, teardownCalledTimes)
}

// TestDerivedSet_inheritMutations tests the inheritMutations method
func TestDerivedSet_inheritMutations(t *testing.T) {
	dSet := newDerivedSet[int]()
	mutations := ds.NewSetMutations[int]().WithAddedElements(newSet(1, 2, 2))

	// Apply mutations
	appliedMutations := dSet.inheritMutations(mutations)
	require.True(t, appliedMutations.AddedElements().Has(1))
	require.True(t, appliedMutations.AddedElements().Has(2))
	require.Equal(t, 2, dSet.Size())
	require.Equal(t, 2, appliedMutations.AddedElements().Size())
	require.Equal(t, 0, appliedMutations.DeletedElements().Size())
}

// TestDerivedSet_applyInheritedMutations tests the applyInheritedMutations method
func TestDerivedSet_applyInheritedMutations(t *testing.T) {
	dSet := newDerivedSet[int]()
	mutations := ds.NewSetMutations[int]().WithAddedElements(newSet(1, 2))

	// Apply mutations
	inheritedMutations, triggerID, _ := dSet.applyInheritedMutations(mutations)
	require.True(t, inheritedMutations.AddedElements().Has(1))
	require.True(t, inheritedMutations.AddedElements().Has(2))
	require.Equal(t, 2, dSet.Size())
	require.NotNil(t, triggerID)
}

// TestSet_Compute tests the Compute method
func TestSet_Compute(t *testing.T) {
	s := newSet[int](1, 2, 3)

	// Define a mutation factory
	mutationFactory := func(set ds.ReadableSet[int]) ds.SetMutations[int] {
		mutations := ds.NewSetMutations[int]()
		if set.Has(2) {
			mutations.WithDeletedElements(newSet(2)) // Delete element 2 if it exists
		}
		mutations.WithAddedElements(newSet(4)) // Add element 4
		return mutations
	}

	// Apply the mutation
	appliedMutations := s.Compute(mutationFactory)

	// Check if the mutations are applied correctly
	require.False(t, s.Has(2), "Element 2 should be deleted")
	require.True(t, s.Has(4), "Element 4 should be added")
	require.Equal(t, 3, s.Size(), "Set size should be 3")
	require.True(t, appliedMutations.DeletedElements().Has(2))
	require.True(t, appliedMutations.AddedElements().Has(4))
	require.Equal(t, 1, appliedMutations.DeletedElements().Size())
	require.Equal(t, 1, appliedMutations.AddedElements().Size())
}

func TestSet_EmptyApply(t *testing.T) {
	appliedMutations := newSet[int](1, 2, 3).Apply(ds.NewSetMutations[int]())

	require.Equal(t, 0, appliedMutations.AddedElements().Size())
	require.Equal(t, 0, appliedMutations.DeletedElements().Size())
	require.True(t, appliedMutations.IsEmpty())
}

// TestSet_Replace tests the Replace method
func TestSet_Replace(t *testing.T) {
	s := newSet[int](1, 2, 3)
	newElements := ds.NewSet[int](4, 5)

	// Replace elements in the set
	removedElements := s.Replace(newElements.ReadOnly())

	// Check if the elements are replaced correctly
	require.False(t, s.Has(1), "Element 1 should be removed")
	require.False(t, s.Has(2), "Element 2 should be removed")
	require.False(t, s.Has(3), "Element 3 should be removed")
	require.True(t, s.Has(4), "Element 4 should be added")
	require.True(t, s.Has(5), "Element 5 should be added")
	require.True(t, removedElements.Has(1))
	require.True(t, removedElements.Has(2))
	require.True(t, removedElements.Has(3))
}

// TestSet_Compute_WithCallback tests the Compute method with a registered callback
func TestSet_Compute_WithCallback(t *testing.T) {
	s := newSet[int](1, 2, 3)
	callbackCalled := false
	updatedElements := ds.NewSet[int]()

	// Register a callback
	unsubscribe := s.OnUpdate(func(appliedMutations ds.SetMutations[int]) {
		callbackCalled = true
		updatedElements = appliedMutations.AddedElements()
	})
	defer unsubscribe()

	// Define a mutation factory
	mutationFactory := func(set ds.ReadableSet[int]) ds.SetMutations[int] {
		mutations := ds.NewSetMutations[int]()
		if set.Has(2) {
			mutations.WithDeletedElements(newSet(2)) // Delete element 2 if it exists
		}
		mutations.WithAddedElements(newSet(4)) // Add element 4
		return mutations
	}

	// Apply the mutation
	s.Compute(mutationFactory)

	// Check if the callback was called and the mutations were applied correctly
	require.True(t, callbackCalled, "OnUpdate callback should be called")
	require.True(t, updatedElements.Has(4), "Element 4 should be added in the callback")
}

// TestSet_Replace_WithCallback tests the Replace method with a registered callback
func TestSet_Replace_WithCallback(t *testing.T) {
	s := newSet[int](1, 2, 3)
	callbackCalled := false
	removedElementsInCallback := ds.NewSet[int]()

	// Register a callback
	unsubscribe := s.OnUpdate(func(appliedMutations ds.SetMutations[int]) {
		callbackCalled = true
		removedElementsInCallback = appliedMutations.DeletedElements()
	})
	defer unsubscribe()

	newElements := ds.NewSet[int](4, 5)

	// Replace elements in the set
	s.Replace(newElements.ReadOnly())

	// Check if the callback was called and the elements are replaced correctly
	require.True(t, callbackCalled, "OnUpdate callback should be called")
	require.True(t, removedElementsInCallback.Has(1), "Element 1 should be removed in the callback")
	require.True(t, removedElementsInCallback.Has(2), "Element 2 should be removed in the callback")
	require.True(t, removedElementsInCallback.Has(3), "Element 3 should be removed in the callback")
}

// TestDerivedSet_inheritMutations_WithCallback tests the inheritMutations method with a registered callback
func TestDerivedSet_inheritMutations_WithCallback(t *testing.T) {
	dSet := newDerivedSet[int]()
	callbackCalled := false
	updatedElements := ds.NewSet[int]()

	// Register a callback
	unsubscribe := dSet.OnUpdate(func(appliedMutations ds.SetMutations[int]) {
		callbackCalled = true
		updatedElements = appliedMutations.AddedElements()
	})
	defer unsubscribe()

	mutations := ds.NewSetMutations[int]().WithAddedElements(newSet(1, 2))

	// Apply mutations
	dSet.inheritMutations(mutations)

	// Check if the callback was called and the mutations were applied correctly
	require.True(t, callbackCalled, "OnUpdate callback should be called")
	require.True(t, updatedElements.Has(1), "Element 1 should be added in the callback")
	require.True(t, updatedElements.Has(2), "Element 2 should be added in the callback")
}

func TestSet_Encoding(t *testing.T) {
	serix.DefaultAPI.RegisterTypeSettings("", serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte))

	testSet := newSet[string]()
	for i := 0; i < 3; i++ {
		testSet.Add(fmt.Sprintf("item%d", i))
	}
	bytes, err := testSet.Encode(serix.DefaultAPI)
	require.NoError(t, err)

	decoded := newSet[string]()
	consumed, err := decoded.Decode(serix.DefaultAPI, bytes)
	require.NoError(t, err)
	require.Equal(t, len(bytes), consumed)

	require.True(t, testSet.Equals(decoded))
}

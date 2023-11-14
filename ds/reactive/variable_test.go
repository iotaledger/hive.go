package reactive

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVariable(t *testing.T) {
	myInt := NewVariable[int]()

	var wg sync.WaitGroup
	wg.Add(2)

	myInt.OnUpdate(func(prevValue, newValue int) {
		if prevValue != 0 {
			require.Equal(t, prevValue+1, newValue)
		}
		// fmt.Println("0> [START] OnUpdate", prevValue, newValue)
		defer fmt.Println("0> [DONE] OnUpdate ", prevValue, newValue)

		time.Sleep(200 * time.Millisecond)

		if newValue == 3 {
			wg.Done()
		}
	})

	go func() {
		time.Sleep(50 * time.Millisecond)

		myInt.OnUpdate(func(prevValue, newValue int) {
			if prevValue != 0 {
				require.Equal(t, prevValue+1, newValue)
			}

			defer fmt.Println("1> [DONE] OnUpdate ", prevValue, newValue)

			time.Sleep(400 * time.Millisecond)

			if newValue == 3 {
				wg.Done()
			}
		})
	}()

	myInt.Set(1)
	myInt.Set(2)
	myInt.Set(3)

	wg.Wait()
}

func TestOnUpdateWithContext(t *testing.T) {
	outerVar := NewVariable[int]()

	innerVars := make([]Variable[int], 0)
	for i := 0; i < 10; i++ {
		innerVars = append(innerVars, NewVariable[int]())
	}

	collectedValues := make([]string, 0)
	unsubscribe := outerVar.OnUpdateWithContext(func(_, monitoredIndex int, withinContext func(subscriptionFactory func() (unsubscribe func()))) {
		withinContext(func() func() {
			return innerVars[monitoredIndex].OnUpdate(func(_, newValue int) {
				collectedValues = append(collectedValues, fmt.Sprintf("%d:%d", monitoredIndex, newValue))
			})
		})
	})

	outerVar.Set(2)

	innerVars[2].Set(3)
	require.Equal(t, []string{"2:3"}, collectedValues)
	innerVars[2].Set(4)
	require.Equal(t, []string{"2:3", "2:4"}, collectedValues)

	outerVar.Set(1)

	innerVars[2].Set(5)
	require.Equal(t, []string{"2:3", "2:4"}, collectedValues)
	innerVars[1].Set(1)
	require.Equal(t, []string{"2:3", "2:4", "1:1"}, collectedValues)
	innerVars[1].Set(2)
	require.Equal(t, []string{"2:3", "2:4", "1:1", "1:2"}, collectedValues)

	unsubscribe()

	innerVars[1].Set(3)
	require.Equal(t, []string{"2:3", "2:4", "1:1", "1:2"}, collectedValues)
}

func TestOnUpdateOnce(t *testing.T) {
	{
		myInt := NewVariable[int]()

		var callCount, calledOldValue, calledNewValue int
		myInt.OnUpdateOnce(func(oldValue, newValue int) {
			callCount++
			calledOldValue = oldValue
			calledNewValue = newValue
		})

		myInt.Set(1)
		require.Equal(t, 1, callCount)
		require.Equal(t, 0, calledOldValue)
		require.Equal(t, 1, calledNewValue)

		myInt.Set(2)
		require.Equal(t, 1, callCount)
		require.Equal(t, 0, calledOldValue)
		require.Equal(t, 1, calledNewValue)
	}

	{
		myInt := NewVariable[int]()

		var callCount, calledOldValue, calledNewValue int
		unsubscribe := myInt.OnUpdateOnce(func(oldValue, newValue int) {
			calledOldValue = oldValue
			calledNewValue = newValue
		})

		unsubscribe()
		require.Equal(t, 0, callCount)
		require.Equal(t, 0, calledOldValue)
		require.Equal(t, 0, calledNewValue)

		myInt.Set(2)
		require.Equal(t, 0, callCount)
		require.Equal(t, 0, calledOldValue)
		require.Equal(t, 0, calledNewValue)
	}
}

// TestVariable_Init tests the Init method of the variable type
func TestVariable_Init(t *testing.T) {
	varValue := 10
	v := newVariable[int]().Init(varValue)
	require.Equal(t, varValue, v.Get())
}

// TestVariable_InheritFrom tests the InheritFrom method of the variable type
func TestVariable_InheritFrom(t *testing.T) {
	parentVar := newVariable[int]().Init(10)
	childVar := newVariable[int]()

	unsubscribe := childVar.InheritFrom(parentVar)
	defer unsubscribe()

	parentVar.Set(20)
	require.Equal(t, 20, childVar.Get())
}

// TestVariable_DeriveValueFrom tests the DeriveValueFrom method of the variable type
func TestVariable_DeriveValueFrom(t *testing.T) {
	sourceVar := newDerivedVariable[int](func(d DerivedVariable[int]) func() { return func() {} }, 10)
	derivedVar := newVariable[int]()

	teardown := derivedVar.DeriveValueFrom(sourceVar)
	defer teardown()

	require.Equal(t, 10, derivedVar.Get())
	sourceVar.Set(20)
	require.Equal(t, 20, derivedVar.Get())
}

// TestVariable_WithValue tests the WithValue method of the variable type
func TestVariable_WithValue(t *testing.T) {
	rVar := newVariable[int]()
	rVar.Set(10)

	setupCalled := false
	teardownCalled := false

	teardown := rVar.WithValue(
		func(value int) (teardown func()) {
			setupCalled = true
			return func() { teardownCalled = true }
		},
	)
	defer teardown()

	rVar.Set(20)
	require.True(t, setupCalled)
	require.True(t, teardownCalled)
}

// TestVariable_WithNonEmptyValue tests the WithNonEmptyValue method of the variable type
func TestVariable_WithNonEmptyValue(t *testing.T) {
	rVar := newVariable[int]()
	rVar.Set(10)
	setupCalled := false

	teardown := rVar.WithNonEmptyValue(
		func(value int) (teardown func()) {
			setupCalled = true
			return func() {}
		},
	)
	defer teardown()

	rVar.Set(20)
	require.True(t, setupCalled)
}

// TestVariable_OnUpdateOnce tests the OnUpdateOnce method of the variable type
func TestVariable_OnUpdateOnce(t *testing.T) {
	rVar := newVariable[int]()
	rVar.Set(10)
	callbackCalled := false

	unsubscribe := rVar.OnUpdateOnce(func(oldValue, newValue int) {
		callbackCalled = true
	})
	defer unsubscribe()

	rVar.Set(20)
	require.True(t, callbackCalled)
}

// TestReadableVariable_OnUpdateOnce_WithCondition tests OnUpdateOnce with a condition
func TestReadableVariable_OnUpdateOnce_WithCondition(t *testing.T) {
	rv := newVariable[int]()
	rv.Set(0)

	var oldValue, newValue int
	conditionNotMet := true

	// Define a condition that is not met
	condition := func(oldVal, newVal int) bool {
		return newVal > 5 // Only trigger if the new value is greater than 5
	}

	// Register the callback with the condition
	unsubscribe := rv.OnUpdateOnce(func(o, n int) {
		oldValue = o
		newValue = n
		conditionNotMet = false
	}, condition)
	defer unsubscribe()

	// Set a value that does not meet the condition
	rv.Set(3)
	require.True(t, conditionNotMet, "Callback should not have been triggered")

	// Set a value that meets the condition
	rv.Set(6)
	require.False(t, conditionNotMet, "Callback should have been triggered")
	require.Equal(t, 3, oldValue)
	require.Equal(t, 6, newValue)
}

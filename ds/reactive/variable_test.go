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

	require.Equal(t, 10, childVar.Get())
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

	var setupElement int
	setupCalledTimes := 0
	var teardownElement int
	teardownCalledTimes := 0

	teardown := rVar.WithValue(
		func(value int) (teardown func()) {
			setupElement = value
			setupCalledTimes++
			return func() {
				teardownElement = value
				teardownCalledTimes++
			}
		},
	)
	defer teardown()

	require.Equal(t, 1, setupCalledTimes)
	require.Equal(t, 0, teardownCalledTimes)
	require.Equal(t, 10, setupElement)
	rVar.Set(20)
	require.Equal(t, 2, setupCalledTimes)
	require.Equal(t, 1, teardownCalledTimes)
	require.Equal(t, 20, setupElement)
	require.Equal(t, 10, teardownElement)
	rVar.Set(0)
	require.Equal(t, 3, setupCalledTimes)
	require.Equal(t, 2, teardownCalledTimes)
	require.Equal(t, 0, setupElement)
	require.Equal(t, 20, teardownElement)
}

// TestVariable_WithNonEmptyValue tests the WithNonEmptyValue method of the variable type
func TestVariable_WithNonEmptyValue(t *testing.T) {
	rVar := newVariable[int]()
	rVar.Set(10)

	var setupElement int
	setupCalledTimes := 0
	var teardownElement int
	teardownCalledTimes := 0

	teardown := rVar.WithNonEmptyValue(
		func(value int) (teardown func()) {
			setupElement = value
			setupCalledTimes++
			return func() {
				teardownElement = value
				teardownCalledTimes++
			}
		},
	)
	defer teardown()

	// The setup and teardown callbacks are only called when the variable is not empty
	require.Equal(t, 1, setupCalledTimes)
	require.Equal(t, 0, teardownCalledTimes)
	require.Equal(t, 10, setupElement)
	rVar.Set(20)
	require.Equal(t, 2, setupCalledTimes)
	require.Equal(t, 1, teardownCalledTimes)
	require.Equal(t, 20, setupElement)
	require.Equal(t, 10, teardownElement)
	rVar.Set(0)
	// The setup is not called for the new empty value, but the teardown for the previous non-empty value is.
	require.Equal(t, 2, setupCalledTimes)
	require.Equal(t, 2, teardownCalledTimes)
	require.Equal(t, 20, setupElement)
	require.Equal(t, 20, teardownElement)
	rVar.Set(0)
	// The setup and teardown callback is not called when the variable is empty
	require.Equal(t, 2, setupCalledTimes)
	require.Equal(t, 2, teardownCalledTimes)
	require.Equal(t, 20, setupElement)
	require.Equal(t, 20, teardownElement)
	rVar.Set(1)
	// The setup is called for the new non-empty value, but the callback for the previous empty value isn't.
	require.Equal(t, 3, setupCalledTimes)
	require.Equal(t, 2, teardownCalledTimes)
	require.Equal(t, 1, setupElement)
	require.Equal(t, 20, teardownElement)
	rVar.Set(2)
	// Both are called
	require.Equal(t, 4, setupCalledTimes)
	require.Equal(t, 3, teardownCalledTimes)
	require.Equal(t, 2, setupElement)
	require.Equal(t, 1, teardownElement)
	rVar.Set(0)
	// Only the teardown is called, as it was non-empty.
	require.Equal(t, 4, setupCalledTimes)
	require.Equal(t, 4, teardownCalledTimes)
	require.Equal(t, 2, setupElement)
	require.Equal(t, 2, teardownElement)
}

// TestVariable_OnUpdateOnce tests the OnUpdateOnce method of the variable type
func TestVariable_OnUpdateOnce(t *testing.T) {
	rVar := newVariable[int]()
	callbackCalledCounter := 0

	rVar.Set(10)

	unsubscribe := rVar.OnUpdateOnce(func(oldValue, newValue int) {
		callbackCalledCounter++
	})
	defer unsubscribe()

	// It should have been already called once, when the callback was registered, as the variable was already updated even
	// before the callback was registered.
	require.Equal(t, 1, callbackCalledCounter)
	rVar.Set(20)
	require.Equal(t, 1, callbackCalledCounter)

	// These should not trigger the callback, as it was only subscribed for a single update.
	rVar.Set(20)
	rVar.Set(1)
	rVar.Set(99)

	require.Equal(t, 1, callbackCalledCounter)
}

// TestReadableVariable_OnUpdateOnce_WithCondition tests OnUpdateOnce with a condition
func TestReadableVariable_OnUpdateOnce_WithCondition(t *testing.T) {
	rv := newVariable[int]()
	rv.Set(0)

	var oldValue, newValue int
	conditionMetCounter := 0

	// Define a condition that is not met
	condition := func(oldVal, newVal int) bool {
		return newVal > 5 // Only trigger if the new value is greater than 5
	}

	// Register the callback with the condition
	unsubscribe := rv.OnUpdateOnce(func(o, n int) {
		oldValue = o
		newValue = n
		conditionMetCounter++
	}, condition)
	defer unsubscribe()

	// Set a value that does not meet the condition
	rv.Set(3)
	require.Equal(t, 0, conditionMetCounter, "Callback should not have been triggered")

	// Set a value that meets the condition
	rv.Set(6)
	require.Equal(t, 1, conditionMetCounter, "Callback should have been triggered only once")
	require.Equal(t, 3, oldValue)
	require.Equal(t, 6, newValue)

	// These meet the condition to update the variable but should not trigger the callback, as it was only subscribed for a single update.
	rv.Set(8)
	rv.Set(11)
	rv.Set(10)
	rv.Set(15)

	require.Equal(t, 1, conditionMetCounter, "Callback should have been triggered only once")
}

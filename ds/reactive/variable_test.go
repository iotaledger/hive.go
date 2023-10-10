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
		//fmt.Println("0> [START] OnUpdate", prevValue, newValue)
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

	innerVars[2].Set(4)
	require.Equal(t, []string{"2:3", "2:4"}, collectedValues)
	innerVars[1].Set(1)
	require.Equal(t, []string{"2:3", "2:4", "1:1"}, collectedValues)
	innerVars[1].Set(2)
	require.Equal(t, []string{"2:3", "2:4", "1:1", "1:2"}, collectedValues)

	unsubscribe()

	innerVars[1].Set(3)
	require.Equal(t, []string{"2:3", "2:4", "1:1", "1:2"}, collectedValues)
}

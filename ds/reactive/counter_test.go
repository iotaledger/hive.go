package reactive_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/ds/reactive"
)

// TestMonitorConcurrency checks if Monitor function correctly updates the counter when multiple goroutines are changing the value
func TestMonitorConcurrency(t *testing.T) {
	//var wg sync.WaitGroup
	condition := func(i int) bool { return i%2 == 0 } // condition is true for even numbers
	c := reactive.NewCounter[int](condition)

	v := reactive.NewVariable[int]()

	// We expect half of the numbers to be even, so the counter should be 500000
	assert.Equal(t, 0, c.Get()) // 0

	unsubscribe := c.Monitor(v)

	assert.Equal(t, 1, c.Get()) // 1

	var wg sync.WaitGroup
	// 1001 goroutines each updating the variable 1001 times (odd number of times)
	for i := 0; i < 1001; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1001; j++ {
				v.Compute(func(currentValue int) int {
					return currentValue + 1
				})
			}
		}()

	}

	wg.Wait()

	// We expect half of the numbers to be even, so the counter should be 500000
	assert.Equal(t, 0, c.Get()) // -1

	// Clean up
	unsubscribe()
}

package reactive_test

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/ds/reactive"
)

// TestCounter_Monitor tests the counter by monitoring a variable concurrently.
func TestCounter_Monitor(t *testing.T) {
	variableCount := 1000

	counter := reactive.NewCounter[int]()
	assert.Equal(t, 0, counter.Get())

	variables := make([]reactive.Variable[int], variableCount)
	for i := 0; i < variableCount; i++ {
		variables[i] = reactive.NewVariable[int]()

		counter.Monitor(variables[i])
	}

	// modify the variables concurrently
	var wg sync.WaitGroup
	for i := 0; i < 1001; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			// choose a random variable and set it to 1 or 0
			variable := variables[rand.Intn(variableCount)]
			variable.Set(rand.Intn(2))
		}()

	}
	wg.Wait()

	trueValues := 0
	for _, variable := range variables {
		if variable.Get() == 1 {
			trueValues++
		}
	}

	assert.Equal(t, trueValues, counter.Get()) // -1
}

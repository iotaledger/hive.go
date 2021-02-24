package timedexecutor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	workerCount = 4
)

func TestTimedExecutor(t *testing.T) {
	timedExecutor := New(workerCount)
	assert.Equal(t, workerCount, timedExecutor.WorkerCount())

	// prepare a list of functions to schedule
	elements := make(map[time.Time]func())
	var expected, actual []int
	now := time.Now().Add(15 * time.Second)

	for i := 0; i < 10; i++ {
		str := i
		elements[now.Add(time.Duration(i)*time.Second)] = func() {
			actual = append(actual, str)
		}
		expected = append(expected, i)
	}

	// insert functions to timedExecutor
	for t, f := range elements {
		timedExecutor.ExecuteAt(f, t)
	}
	assert.Equal(t, len(elements), timedExecutor.Size())

	assert.Eventually(t, func() bool { return len(actual) == len(expected) }, 5*time.Minute, 100*time.Millisecond)
	assert.Equal(t, 0, timedExecutor.Size())
	assert.ElementsMatch(t, expected, actual)

	timedExecutor.Shutdown()
}

package async

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func slowFunc() {
	time.Sleep(100 * time.Millisecond)
}

func TestNonBlockingQueue_Queuing(t *testing.T) {
	wp := NonBlockingQueueWorkerPool{}
	capacity := runtime.GOMAXPROCS(0)

	assert.Equal(t, wp.Capacity(), capacity)

	for i := 0; i < capacity*3-1; i++ {
		assert.True(t, wp.Submit(slowFunc))
	}

	enqueuedAndExecuted := false
	assert.True(t, wp.Submit(func() {
		enqueuedAndExecuted = true
	}))

	// Reached queue capacity
	assert.False(t, wp.Submit(slowFunc))
	assert.False(t, wp.Submit(slowFunc))

	assert.False(t, enqueuedAndExecuted)
	wp.tasksWg.Wait()
	assert.True(t, enqueuedAndExecuted)
}

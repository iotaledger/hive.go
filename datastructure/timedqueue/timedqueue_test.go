package datastructure

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimedQueue_Size(t *testing.T) {
	timedQueue := New()

	timedQueue.Add(1, time.Now())
	assert.Equal(t, 1, timedQueue.Size())

	timedQueue.Add(2, time.Now())
	assert.Equal(t, 2, timedQueue.Size())

	timedQueue.Add(3, time.Now())
	assert.Equal(t, 3, timedQueue.Size())

	timedQueue.Poll(false)
	assert.Equal(t, 2, timedQueue.Size())

	timedQueue.Shutdown(CancelPendingElements)
	assert.Equal(t, 0, timedQueue.Size())
}

func TestTimedQueue_Add(t *testing.T) {
	timedQueue := New()

	element := timedQueue.Add(1, time.Now())
	assert.Equal(t, 1, timedQueue.Size())

	element.Cancel()
	assert.Equal(t, 0, timedQueue.Size())
}

func TestTimedQueue_Poll(t *testing.T) {
	timedQueue := New()

	seenElements := make(map[int]bool)

	var waitForShutdownWg sync.WaitGroup
	waitForShutdownWg.Add(1)
	go func() {
		for currentEntry := timedQueue.Poll(true); currentEntry != nil; currentEntry = timedQueue.Poll(true) {
			seenElements[currentEntry.(int)] = true
		}

		waitForShutdownWg.Done()
	}()

	timedQueue.Add(2, time.Now().Add(1*time.Second))
	elem := timedQueue.Add(4, time.Now().Add(2*time.Second))
	timedQueue.Add(6, time.Now().Add(3*time.Second))

	time.Sleep(1 * time.Second)

	elem.Cancel()

	time.Sleep(2 * time.Second)

	timedQueue.Add(7, time.Now().Add(time.Second))
	timedQueue.Add(8, time.Now().Add(4*time.Second))

	timedQueue.Shutdown(IgnorePendingTimeouts | PanicOnModificationsAfterShutdown)

	waitForShutdownWg.Wait()

	assert.Equal(t, map[int]bool{
		2: true,
		6: true,
		7: true,
		8: true,
	}, seenElements)
}

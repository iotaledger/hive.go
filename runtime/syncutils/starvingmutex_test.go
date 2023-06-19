//nolint:staticcheck // we don't care about these linters in test cases
package syncutils

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/runtime/debug"
)

func Benchmark(b *testing.B) {
	var wg sync.WaitGroup

	mutex := NewStarvingMutex()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 20; j++ {
			wg.Add(1)
			go func(goRoutineID int) {
				if goRoutineID%2 == 0 {
					mutex.Lock()
					mutex.Unlock()
				} else {
					mutex.RLock()
					mutex.RUnlock()
				}
				wg.Done()
			}(j)
		}
	}

	wg.Wait()
}

func TestStarvingMutex(t *testing.T) {
	debug.SetEnabled(false)
	debug.DeadlockDetectionTimeout = 100 * time.Millisecond

	mutex := NewStarvingMutex()

	mutex.RLock()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		mutex.Lock()
		mutex.Unlock()
		wg.Done()
	}()

	time.Sleep(500 * time.Millisecond)

	mutex.RLock()
	mutex.RLock()

	assert.Equal(t, false, mutex.writerActive)
	assert.Equal(t, 3, mutex.readersActive)

	mutex.RUnlock()
	mutex.RUnlock()
	mutex.RUnlock()

	time.Sleep(500 * time.Millisecond)

	mutex.Lock()
	mutex.Unlock()

	wg.Wait()

	assert.Equal(t, false, mutex.writerActive)
	assert.Equal(t, 0, mutex.readersActive)
	assert.Equal(t, 0, mutex.pendingWriters)

	mutex.Lock()

	assert.Equal(t, true, mutex.writerActive)

	go func() {
		mutex.RLock()
		assert.Equal(t, false, mutex.writerActive)
		assert.Equal(t, 1, mutex.readersActive)
		mutex.RUnlock()
	}()

	go func() {
		time.Sleep(1 * time.Second)
		assert.Equal(t, true, mutex.writerActive)
		assert.Equal(t, 0, mutex.readersActive)
		mutex.Unlock()
	}()

	mutex.Lock()
	assert.Equal(t, true, mutex.writerActive)
	assert.Equal(t, 0, mutex.readersActive)
	mutex.Unlock()

	assert.Equal(t, false, mutex.writerActive)
	assert.Equal(t, 0, mutex.readersActive)
	assert.Equal(t, 0, mutex.pendingWriters)
}

func TestStarvingMutexParallel(t *testing.T) {
	const count = 10000
	mutex := NewStarvingMutex()

	// RLock
	{
		var wg sync.WaitGroup
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func() {
				mutex.RLock()
				wg.Done()
			}()
		}
		wg.Wait()

		assert.Equal(t, false, mutex.writerActive)
		assert.Equal(t, count, mutex.readersActive)
		assert.Equal(t, 0, mutex.pendingWriters)
	}

	// RUnlock
	{
		var wg sync.WaitGroup
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func() {
				mutex.RUnlock()
				wg.Done()
			}()
		}
		wg.Wait()

		assert.Equal(t, false, mutex.writerActive)
		assert.Equal(t, 0, mutex.readersActive)
		assert.Equal(t, 0, mutex.pendingWriters)
	}

	// Lock / Unlock
	{
		var wg sync.WaitGroup
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func() {
				mutex.Lock()
				assert.Equal(t, true, mutex.writerActive)
				assert.Equal(t, 0, mutex.readersActive)
				mutex.Unlock()
				wg.Done()
			}()
		}
		wg.Wait()

		assert.Equal(t, false, mutex.writerActive)
		assert.Equal(t, 0, mutex.readersActive)
		assert.Equal(t, 0, mutex.pendingWriters)
	}
}

func TestStarvingMutexParallelWithLock(t *testing.T) {
	const count = 100
	mutex := NewStarvingMutex()

	// RLock
	{
		var wg sync.WaitGroup
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func() {
				mutex.RLock()
				wg.Done()
			}()
		}
		wg.Wait()

		assert.Equal(t, false, mutex.writerActive)
		assert.Equal(t, count, mutex.readersActive)
		assert.Equal(t, 0, mutex.pendingWriters)
	}

	// RUnlock / Lock
	{
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			mutex.Lock()
			wg.Done()
		}()

		wg.Add(count)
		for i := 0; i < count; i++ {
			go func() {
				mutex.RUnlock()
				wg.Done()
			}()
		}
		wg.Wait()

		assert.Equal(t, true, mutex.writerActive)
		assert.Equal(t, 0, mutex.readersActive)
		assert.Equal(t, 0, mutex.pendingWriters)
	}

	// Unlock
	mutex.Unlock()
	assert.Equal(t, false, mutex.writerActive)
	assert.Equal(t, 0, mutex.readersActive)
	assert.Equal(t, 0, mutex.pendingWriters)
}

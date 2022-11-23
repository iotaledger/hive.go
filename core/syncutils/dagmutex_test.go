package syncutils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DAGMutex(t *testing.T) {
	mutex := NewDAGMutex[int]()

	mutex.RLock(1, 2)
	_, exists := mutex.mutexes.Get(1)
	assert.True(t, exists)
	count, _ := mutex.consumerCounter.Get(1)
	assert.Equal(t, 1, count)

	_, exists = mutex.mutexes.Get(2)
	assert.True(t, exists)
	count, _ = mutex.consumerCounter.Get(2)
	assert.Equal(t, 1, count)

	mutex.RLock(1, 4)
	_, exists = mutex.mutexes.Get(1)
	assert.True(t, exists)
	count, _ = mutex.consumerCounter.Get(1)
	assert.Equal(t, 2, count)

	_, exists = mutex.mutexes.Get(2)
	assert.True(t, exists)
	count, _ = mutex.consumerCounter.Get(2)
	assert.Equal(t, 1, count)

	_, exists = mutex.mutexes.Get(4)
	assert.True(t, exists)
	count, _ = mutex.consumerCounter.Get(4)
	assert.Equal(t, 1, count)

	mutex.RLock(1)
	_, exists = mutex.mutexes.Get(1)
	assert.True(t, exists)
	count, _ = mutex.consumerCounter.Get(1)
	assert.Equal(t, 3, count)

	_, exists = mutex.mutexes.Get(2)
	assert.True(t, exists)
	count, _ = mutex.consumerCounter.Get(2)
	assert.Equal(t, 1, count)

	_, exists = mutex.mutexes.Get(4)
	assert.True(t, exists)
	count, _ = mutex.consumerCounter.Get(4)
	assert.Equal(t, 1, count)

	mutex.RUnlock(1, 4)
	_, exists = mutex.mutexes.Get(1)
	assert.True(t, exists)
	count, _ = mutex.consumerCounter.Get(1)
	assert.Equal(t, 2, count)

	_, exists = mutex.mutexes.Get(2)
	assert.True(t, exists)
	count, _ = mutex.consumerCounter.Get(2)
	assert.Equal(t, 1, count)

	_, exists = mutex.mutexes.Get(4)
	assert.False(t, exists)
	count, _ = mutex.consumerCounter.Get(4)
	assert.Equal(t, 0, count)

	mutex.RUnlock(1, 2)
	mutex.RUnlock(1)
	_, exists = mutex.mutexes.Get(1)
	assert.False(t, exists)
	count, _ = mutex.consumerCounter.Get(1)
	assert.Equal(t, 0, count)

	_, exists = mutex.mutexes.Get(2)
	assert.False(t, exists)
	count, _ = mutex.consumerCounter.Get(2)
	assert.Equal(t, 0, count)

	mutex.Lock(1)
	_, exists = mutex.mutexes.Get(1)
	assert.True(t, exists)
	count, _ = mutex.consumerCounter.Get(1)
	assert.Equal(t, 1, count)

	mutex.Unlock(1)
	_, exists = mutex.mutexes.Get(1)
	assert.False(t, exists)
	count, _ = mutex.consumerCounter.Get(1)
	assert.Equal(t, 0, count)
}

func TestDAGMutexParallel(t *testing.T) {
	const count = 10000
	mutex := NewDAGMutex[int]()

	// RLock a bunch of things in parallel
	{
		var wg sync.WaitGroup
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func(i int) {
				mutex.RLock(i, i+1)
				wg.Done()
			}(i)
		}

		wg.Add(count)
		for i := 0; i < count; i++ {
			go func(i int) {
				mutex.RLock(i, i+1, i+10)
				wg.Done()
			}(i)
		}
		wg.Wait()
	}

	// Lock a bunch of things in parallel and RUnlock
	{
		var wg sync.WaitGroup
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func(i int) {
				mutex.Lock(i)
				wg.Done()
			}(i)
		}
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func(i int) {
				mutex.RUnlock(i, i+1)
				wg.Done()
			}(i)
		}

		wg.Add(count)
		for i := 0; i < count; i++ {
			go func(i int) {
				mutex.RUnlock(i, i+1, i+10)
				wg.Done()
			}(i)
		}
		wg.Wait()
	}

	// Unlock everything in parallel
	{
		var wg sync.WaitGroup
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func(i int) {
				mutex.Unlock(i)
				wg.Done()
			}(i)
		}
		wg.Wait()
	}

	// All entities' locks should be empty
	assert.True(t, mutex.consumerCounter.IsEmpty())
	assert.True(t, mutex.mutexes.IsEmpty())
}

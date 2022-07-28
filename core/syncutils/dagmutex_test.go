package syncutils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DAGMutex(t *testing.T) {
	mutex := NewDAGMutex[int]()

	mutex.RLock(1, 2)
	assert.NotNil(t, mutex.mutexes[1])
	assert.Equal(t, 1, mutex.consumerCounter[1])
	assert.NotNil(t, mutex.mutexes[2])
	assert.Equal(t, 1, mutex.consumerCounter[2])

	mutex.RLock(1, 4)
	assert.NotNil(t, mutex.mutexes[1])
	assert.Equal(t, 2, mutex.consumerCounter[1])
	assert.NotNil(t, mutex.mutexes[2])
	assert.Equal(t, 1, mutex.consumerCounter[2])
	assert.NotNil(t, mutex.mutexes[4])
	assert.Equal(t, 1, mutex.consumerCounter[4])

	mutex.RLock(1)
	assert.NotNil(t, mutex.mutexes[1])
	assert.Equal(t, 3, mutex.consumerCounter[1])
	assert.NotNil(t, mutex.mutexes[2])
	assert.Equal(t, 1, mutex.consumerCounter[2])
	assert.NotNil(t, mutex.mutexes[4])
	assert.Equal(t, 1, mutex.consumerCounter[4])

	mutex.RUnlock(1, 4)
	assert.NotNil(t, mutex.mutexes[1])
	assert.Equal(t, 2, mutex.consumerCounter[1])
	assert.NotNil(t, mutex.mutexes[2])
	assert.Equal(t, 1, mutex.consumerCounter[2])
	assert.Nil(t, mutex.mutexes[4])
	assert.Equal(t, 0, mutex.consumerCounter[4])

	mutex.RUnlock(1, 2)
	mutex.RUnlock(1)
	assert.Nil(t, mutex.mutexes[1])
	assert.Equal(t, 0, mutex.consumerCounter[1])
	assert.Nil(t, mutex.mutexes[2])
	assert.Equal(t, 0, mutex.consumerCounter[2])

	mutex.Lock(1)
	assert.NotNil(t, mutex.mutexes[1])
	assert.Equal(t, 1, mutex.consumerCounter[1])
	mutex.Unlock(1)

	assert.Nil(t, mutex.mutexes[1])
	assert.Equal(t, 0, mutex.consumerCounter[1])
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
	assert.Empty(t, mutex.consumerCounter)
	assert.Empty(t, mutex.mutexes)
}

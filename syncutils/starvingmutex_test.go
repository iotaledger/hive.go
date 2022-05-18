package syncutils

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/debug"
)

func Benchmark(b *testing.B) {
	var wg sync.WaitGroup

	mutex := NewStarvingMutex()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 20; j++ {
			wg.Add(1)
			go func(goRoutineId int) {
				if i%2 == 0 {
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

func Test(t *testing.T) {
	debug.SetEnabled(true)
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

	mutex.RLock()
	mutex.RLock()
	mutex.RUnlock()
	mutex.RUnlock()
	mutex.RUnlock()

	time.Sleep(500 * time.Millisecond)

	mutex.Lock()
	mutex.Unlock()

	wg.Wait()

	assert.Equal(t, false, mutex.writerActive)
	assert.Equal(t, 0, mutex.readersActive)

	mutex.Lock()

	go func() {
		mutex.RLock()
		mutex.RUnlock()
	}()

	go func() {
		time.Sleep(1 * time.Second)
		mutex.Unlock()
	}()

	mutex.Lock()
	mutex.Unlock()
}

package syncutils

import (
	"sync"
	"testing"
	"time"
)

func Test_DAGMutex(t *testing.T) {
	mutex := NewDAGMutex[int]()

	mutex.RLock(1, 2)
	defer mutex.RUnlock(1, 2)

	mutex.RLock(1, 4)
	defer mutex.RUnlock(1, 4)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		mutex.Lock(1)
		mutex.Unlock(1)
	}()

	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	mutex.RLock(1)
	defer mutex.RUnlock(1)
}

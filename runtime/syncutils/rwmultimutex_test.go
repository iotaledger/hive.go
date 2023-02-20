package syncutils

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func BenchmarkRWMultiMutex_Lock(b *testing.B) {
	var mutex RWMultiMutex

	for i := 0; i < b.N; i++ {
		mutex.Lock(1)
		mutex.Unlock(1)
	}
}

func TestRWMultiMutex_RLock(t *testing.T) {
	var m RWMultiMutex

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		m.RLock(7)
		fmt.Println("READ: 7")
		time.Sleep(2 * time.Second)
		m.RUnlock(7)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		m.RLock(1, 7)
		fmt.Println("READ: 1, 7")
		time.Sleep(2 * time.Second)
		m.RUnlock(1, 7)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		m.Lock(7)
		fmt.Println("WRITE: 7")
		time.Sleep(2 * time.Second)
		m.Unlock(7)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		m.RLock(1)
		fmt.Println("READ: 1")
		time.Sleep(2 * time.Second)
		m.RUnlock(1)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		m.Lock(3, 2, 1)
		fmt.Println("WRITE: 3, 2, 1")
		time.Sleep(2 * time.Second)
		m.Unlock(3, 2, 1)
		wg.Done()
	}()

	wg.Wait()
}

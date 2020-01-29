package syncutils

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func BenchmarkMultiMutex(b *testing.B) {
	var mutex MultiMutex

	for i := 0; i < b.N; i++ {
		mutex.Lock(i)
		mutex.Unlock(i)
	}
}

func BenchmarkMutex(b *testing.B) {
	var mutex sync.Mutex

	for i := 0; i < b.N; i++ {
		mutex.Lock()
		mutex.Unlock()
	}
}

func BenchmarkMutex_Parallel(b *testing.B) {
	var mutex sync.Mutex

	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			mutex.Lock()
			mutex.Unlock()

			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkMultiMutex_Parallel(b *testing.B) {
	var mutex MultiMutex

	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		var x = i

		wg.Add(1)
		go func() {
			mutex.Lock(x)
			mutex.Unlock(x)

			wg.Done()
		}()
	}

	wg.Wait()
}

func TestMultiMutex_Lock(t *testing.T) {
	m := NewMultiMutex()

	const (
		transactionStorageMutex = iota
		metadataStorageMutex
	)

	var mutex MultiMutex

	mutex.Lock(transactionStorageMutex, metadataStorageMutex)
	fmt.Println("Test")
	mutex.Unlock(metadataStorageMutex, transactionStorageMutex)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		m.Lock(7)
		fmt.Println("7")
		time.Sleep(2 * time.Second)
		m.Unlock(7)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		m.Lock(1, 2)
		fmt.Println("1, 2")
		time.Sleep(2 * time.Second)
		m.Unlock(1, 2)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		m.Lock(3)
		fmt.Println("3")
		time.Sleep(2 * time.Second)
		m.Unlock(3)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		m.Lock(1)
		fmt.Println("1")
		time.Sleep(2 * time.Second)
		m.Unlock(1)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		m.Lock(3, 2, 1)
		fmt.Println("3, 2, 1")
		time.Sleep(2 * time.Second)
		m.Unlock(3, 2, 1)
		wg.Done()
	}()

	wg.Wait()
}

func TestMultiMutex_LockNested(t *testing.T) {
	const (
		some = iota
		someOther
	)

	var wg sync.WaitGroup

	var mutex MultiMutex

	doSth := func() {
		mutex.Lock(some, someOther)

		// ... do sth
		time.Sleep(100 * time.Millisecond)

		mutex.Unlock(some, someOther) // swapping the order to (someOther, some) is also fine

		wg.Done()
	}

	doSthElse := func() {
		mutex.Lock(someOther)

		// ... do sth
		time.Sleep(100 * time.Millisecond)

		if true {
			mutex.Lock(some)

			// ... do sth else
			time.Sleep(100 * time.Millisecond)

			mutex.Unlock(some)
		}

		mutex.Unlock(someOther)

		wg.Done()
	}

	wg.Add(2)

	go doSthElse()
	go doSth()

	wg.Wait()
}

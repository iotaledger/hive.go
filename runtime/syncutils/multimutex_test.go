//nolint:staticcheck // we don't care about these linters in test cases
package syncutils

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func BenchmarkMultiMutex(b *testing.B) {
	mutex := NewMultiMutex()

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
	mutex := NewMultiMutex()

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

	mutex := NewMultiMutex()

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

	mutex := NewMultiMutex()

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

func TestMultiMutex_Lock2(t *testing.T) {
	mutex := NewMultiMutex()

	acquireLockAndPrint(mutex, 1, 3)
	acquireLockAndPrint(mutex, 2)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		acquireLockAndPrint(mutex, 2, 3)
		wg.Done()
	}()

	fmt.Println("Waiting...")
	time.Sleep(time.Second)

	mutex.Unlock(2, 3)

	wg.Wait()
}

func acquireLockAndPrint(m *MultiMutex, ids ...interface{}) {
	m.Lock(ids...)
	// fmt.Println("Locked", ids)
}

func TestMultiMutexMassiveParallel(t *testing.T) {
	mutex := NewMultiMutex()

	N := 20000
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			// Access L random locks.
			L := 100
			ids := make([]interface{}, 0, L)
			for _, x := range rand.Perm(L) {
				ids = append(ids, x)
			}
			acquireLockAndPrint(mutex, ids...)

			// work
			time.Sleep(100 * time.Nanosecond)
			mutex.Unlock(ids...)

			wg.Done()
		}()
	}

	wg.Wait()

	assert.Equal(t, 0, len(mutex.locks))
}

func BenchmarkMultiMutexMassiveParallel(b *testing.B) {
	threads := 4 * runtime.GOMAXPROCS(0)

	ids := make([][][]interface{}, b.N)
	for i := 0; i < b.N; i++ {
		ids[i] = make([][]interface{}, threads)
		for t := 0; t < threads; t++ {
			// Access L random locks.
			L := 50
			ids[i][t] = make([]interface{}, L)
			for j, x := range rand.Perm(L) {
				ids[i][t][j] = x
			}
		}
	}
	b.ResetTimer()

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		mutex := NewMultiMutex()
		wg.Add(threads)
		for t := 0; t < threads; t++ {
			go func(i, t int) {

				mutex.Lock(ids[i][t][:10]...)
				// work
				// time.Sleep(100 * time.Nanosecond)
				mutex.Unlock(ids[i][t][:10]...)

				wg.Done()
			}(i, t)
		}

	}
	wg.Wait()
}

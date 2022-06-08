package syncutils

import (
	"fmt"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/debug"
	"github.com/iotaledger/hive.go/stringify"
	"github.com/iotaledger/hive.go/types"
)

// A StarvingMutex is a reader/writer mutual exclusion lock.
// The lock can be held by an arbitrary number of readers or a single writer.
// The zero value for a StarvingMutex is an unlocked mutex.
//
// A StarvingMutex must not be copied after first use.
//
// If a goroutine holds a StarvingMutex for reading and another goroutine might
// call Lock, other goroutine can acquire a read lock . This allows
// recursive read locking. However, this can result in starvation of goroutine
// that tried to acquire write lock on the mutex.
//  A blocked Lock call does not exclude new readers from acquiring the lock.
type StarvingMutex struct {
	readersActive  int
	writerActive   bool
	pendingWriters int

	mutex      sync.Mutex
	readerCond sync.Cond
	writerCond sync.Cond
}

// NewStarvingMutex creates a new StarvingMutex.
func NewStarvingMutex() *StarvingMutex {
	fm := &StarvingMutex{}
	fm.readerCond.L = &fm.mutex
	fm.writerCond.L = &fm.mutex
	return fm
}

// RLock locks starving mutex for reading.
//
// It should not be used for recursive read locking; a blocked Lock
// call excludes new readers from acquiring the lock. See the
// documentation on the RWMutex type.
func (f *StarvingMutex) RLock() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	var doneChan chan types.Empty
	if debug.GetEnabled() {
		doneChan = make(chan types.Empty, 1)

		go f.detectDeadlock("RLock", debug.CallerStackTrace(), doneChan)
	}

	for f.writerActive {
		f.readerCond.Wait()
	}

	if debug.GetEnabled() {
		close(doneChan)
	}

	f.readersActive++
}

// RUnlock undoes a single RLock call;
// it does not affect other simultaneous readers.
// It is a run-time error if mutex is not locked for reading
// on entry to RUnlock.
func (f *StarvingMutex) RUnlock() {
	f.mutex.Lock()

	if f.readersActive == 0 {
		panic("RUnlock called without RLock")
	}

	if f.writerActive {
		panic("RUnlock called while writer active")
	}

	f.readersActive--

	if f.readersActive == 0 && f.pendingWriters > 0 {
		f.mutex.Unlock()
		f.writerCond.Signal()
		return
	}
	f.mutex.Unlock()
}

// Lock locks starving mutex for writing.
// If the lock is already locked for reading or writing,
// Lock blocks until the lock is available.
func (f *StarvingMutex) Lock() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	var doneChan chan types.Empty
	if debug.GetEnabled() {
		doneChan = make(chan types.Empty, 1)

		go f.detectDeadlock("Lock", debug.CallerStackTrace(), doneChan)
	}

	f.pendingWriters++
	for !f.canWrite() {
		f.writerCond.Wait()
	}
	if debug.GetEnabled() {
		close(doneChan)
	}
	f.pendingWriters--
	f.writerActive = true
}

// Unlock unlocks starving mutex for writing. It is a run-time error if mutex is
// not locked for writing on entry to Unlock.
//
// As with Mutexes, a locked StarvingMutex is not associated with a particular
// goroutine. One goroutine may RLock (Lock) a StarvingMutex and then
// arrange for another goroutine to RUnlock (Unlock) it.
func (f *StarvingMutex) Unlock() {
	f.mutex.Lock()

	if f.readersActive > 0 {
		panic("Unlock called while readers active")
	}

	f.writerActive = false
	if f.pendingWriters == 0 {
		f.mutex.Unlock()
		f.readerCond.Broadcast()
		return
	}

	f.mutex.Unlock()
	f.writerCond.Signal()
}

// String returns a string representation of the StarvingMutex.
func (f *StarvingMutex) String() (humanReadable string) {
	return stringify.Struct("StarvingMutex",
		stringify.StructField("WriterActive", f.writerActive),
		stringify.StructField("ReadersActive", f.readersActive),
		stringify.StructField("PendingWriters", f.pendingWriters),
	)
}

func (f *StarvingMutex) canWrite() bool {
	return !f.writerActive && f.readersActive == 0
}

func (f *StarvingMutex) detectDeadlock(lockType string, trace string, done chan types.Empty) {
	select {
	case <-done:
		return
	case <-time.After(debug.DeadlockDetectionTimeout):
		fmt.Println("possible deadlock while trying to acquire " + lockType + " (" + debug.DeadlockDetectionTimeout.String() + ") ...")
		fmt.Println("\n" + trace)
	}
}

package syncutils

import (
	"fmt"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/stringify"
)

type StarvingMutex struct {
	readersActive  int
	writerActive   bool
	pendingWriters int

	mutex sync.Mutex

	readerCond sync.Cond
	writerCond sync.Cond
}

func NewStarvingMutex() *StarvingMutex {
	fm := &StarvingMutex{}
	fm.readerCond.L = &fm.mutex
	fm.writerCond.L = &fm.mutex
	return fm
}

func (f *StarvingMutex) RLock() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	waiting := true
	go func() {
		time.Sleep(10 * time.Second)
		if waiting {
			fmt.Println("DEADLOCK!!!")
		}
	}()

	for f.writerActive {
		f.readerCond.Wait()
	}
	waiting = false

	f.readersActive++
}

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

func (f *StarvingMutex) Lock() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	waiting := true
	go func() {
		time.Sleep(10 * time.Second)
		if waiting {
			fmt.Println("DEADLOCK!!!")
		}
	}()

	f.pendingWriters++
	for !f.canWrite() {
		f.writerCond.Wait()
	}
	waiting = false
	f.pendingWriters--
	f.writerActive = true
}

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

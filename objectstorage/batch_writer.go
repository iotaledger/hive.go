package objectstorage

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/iotaledger/goshimmer/packages/typeutils"

	"github.com/dgraph-io/badger"

	"github.com/iotaledger/goshimmer/packages/daemon"
)

var writeWg sync.WaitGroup

var timeoutWg sync.WaitGroup

var waitingForTimeout bool

var startStopMutex sync.Mutex

var running int32 = 0

var batchQueue = make(chan *CachedObject, BATCH_WRITER_QUEUE_SIZE)

func StartBatchWriter() {
	startStopMutex.Lock()
	if atomic.LoadInt32(&running) == 0 {
		atomic.StoreInt32(&running, 1)

		go runBatchWriter()
	}
	startStopMutex.Unlock()
}

func StopBatchWriter() {
	startStopMutex.Lock()
	if atomic.LoadInt32(&running) != 0 {
		atomic.StoreInt32(&running, 0)

		writeWg.Wait()
	}
	startStopMutex.Unlock()
}

func WaitForWritesToFlush() {
	timeoutWg.Wait()
}

func batchWrite(object *CachedObject) {
	if atomic.LoadInt32(&running) == 0 {
		StartBatchWriter()
	}

	batchQueue <- object
}

func writeObject(writeBatch *badger.WriteBatch, cachedObject *CachedObject) {
	objectStorage := cachedObject.objectStorage

	if consumers := atomic.LoadInt32(&(cachedObject.consumers)); consumers == 0 && atomic.AddInt32(&(cachedObject.stored), 1) == 1 {
		if cachedObject.IsDeleted() {
			if err := writeBatch.Delete(objectStorage.generatePrefix([][]byte{cachedObject.key})); err != nil {
				panic(err)
			}
		} else if atomic.LoadInt32(&(cachedObject.store)) == 1 && cachedObject.value != nil {
			marshaledObject, _ := cachedObject.value.MarshalBinary()

			if err := writeBatch.Set(objectStorage.generatePrefix([][]byte{cachedObject.value.GetStorageKey()}), marshaledObject); err != nil {
				panic(err)
			}
		}
	} else if consumers < 0 {
		panic("too many unregistered consumers of cached object")
	}
}

func releaseObject(cachedObject *CachedObject) {
	objectStorage := cachedObject.objectStorage

	objectStorage.cacheMutex.Lock()
	if consumers := atomic.LoadInt32(&(cachedObject.consumers)); consumers == 0 {
		delete(objectStorage.cachedObjects, typeutils.BytesToString(cachedObject.key))
	}
	objectStorage.cacheMutex.Unlock()
}

func runBatchWriter() {
	badgerInstance := GetBadgerInstance()

	for atomic.LoadInt32(&running) == 1 {
		writeWg.Add(1)

		if !waitingForTimeout {
			waitingForTimeout = true
			timeoutWg.Add(1)
		}

		wb := badgerInstance.NewWriteBatch()

		writtenValues := make([]*CachedObject, BATCH_WRITER_BATCH_SIZE)
		writtenValuesCounter := 0
	COLLECT_VALUES:
		for writtenValuesCounter < BATCH_WRITER_BATCH_SIZE {
			select {
			case objectToPersist := <-batchQueue:
				writeObject(wb, objectToPersist)

				writtenValues[writtenValuesCounter] = objectToPersist
				writtenValuesCounter++
			case <-time.After(BATCH_WRITER_BATCH_TIMEOUT):
				waitingForTimeout = false
				timeoutWg.Done()

				break COLLECT_VALUES
			case <-daemon.ShutdownSignal:
				break COLLECT_VALUES
			}
		}

		if err := wb.Flush(); err != nil && err != badger.ErrBlockedWrites {
			panic(err)
		}

		for _, cachedObject := range writtenValues {
			if cachedObject != nil {
				releaseObject(cachedObject)
			}
		}

		writeWg.Done()
	}
}

const (
	BATCH_WRITER_QUEUE_SIZE    = BATCH_WRITER_BATCH_SIZE
	BATCH_WRITER_BATCH_SIZE    = 1024
	BATCH_WRITER_BATCH_TIMEOUT = 500 * time.Millisecond
)

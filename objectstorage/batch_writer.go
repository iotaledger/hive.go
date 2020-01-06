package objectstorage

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/badger/v2"

	"github.com/iotaledger/hive.go/syncutils"
	"github.com/iotaledger/hive.go/typeutils"
)

const (
	BATCH_WRITER_QUEUE_SIZE    = BATCH_WRITER_BATCH_SIZE
	BATCH_WRITER_BATCH_SIZE    = 1024
	BATCH_WRITER_BATCH_TIMEOUT = 500 * time.Millisecond
)

type BatchedWriter struct {
	badgerInstance    *badger.DB
	writeWg           sync.WaitGroup
	timeoutWg         sync.WaitGroup
	waitingForTimeout bool
	startStopMutex    syncutils.Mutex
	running           int32
	batchQueue        chan *CachedObject
}

func NewBatchedWriter(badgerInstance *badger.DB) *BatchedWriter {
	return &BatchedWriter{
		badgerInstance:    badgerInstance,
		writeWg:           sync.WaitGroup{},
		timeoutWg:         sync.WaitGroup{},
		waitingForTimeout: false,
		startStopMutex:    syncutils.Mutex{},
		running:           0,
		batchQueue:        make(chan *CachedObject, BATCH_WRITER_QUEUE_SIZE),
	}
}

func (bw *BatchedWriter) StartBatchWriter() {
	bw.startStopMutex.Lock()
	if atomic.LoadInt32(&bw.running) == 0 {
		atomic.StoreInt32(&bw.running, 1)

		go bw.runBatchWriter()
	}
	bw.startStopMutex.Unlock()
}

func (bw *BatchedWriter) StopBatchWriter() {
	bw.startStopMutex.Lock()
	if atomic.LoadInt32(&bw.running) != 0 {
		atomic.StoreInt32(&bw.running, 0)

		bw.writeWg.Wait()
	}
	bw.startStopMutex.Unlock()
}

func (bw *BatchedWriter) WaitForWritesToFlush() {
	bw.timeoutWg.Wait()
}

func (bw *BatchedWriter) batchWrite(object *CachedObject) {
	if atomic.LoadInt32(&bw.running) == 0 {
		bw.StartBatchWriter()
	}

	bw.batchQueue <- object
}

func (bw *BatchedWriter) writeObject(writeBatch *badger.WriteBatch, cachedObject *CachedObject) {
	objectStorage := cachedObject.objectStorage

	if consumers := atomic.LoadInt32(&(cachedObject.consumers)); consumers == 0 {
		if storableObject := cachedObject.Get(); storableObject != nil {
			if storableObject.IsDeleted() {
				if err := writeBatch.Delete(objectStorage.generatePrefix([][]byte{cachedObject.key})); err != nil {
					panic(err)
				}
			} else if storableObject.PersistenceEnabled() && storableObject.IsModified() {
				storableObject.SetModified(false)

				marshaledObject, _ := storableObject.MarshalBinary()
				if err := writeBatch.Set(objectStorage.generatePrefix([][]byte{storableObject.GetStorageKey()}), marshaledObject); err != nil {
					panic(err)
				}
			}
		} else if cachedObject.blindDelete.IsSet() {
			if err := writeBatch.Delete(objectStorage.generatePrefix([][]byte{cachedObject.key})); err != nil {
				panic(err)
			}
		}
	} else if consumers < 0 {
		panic("too many unregistered consumers of cached object")
	}
}

func (bw *BatchedWriter) releaseObject(cachedObject *CachedObject) {
	objectStorage := cachedObject.objectStorage

	objectStorage.cacheMutex.Lock()
	if consumers := atomic.LoadInt32(&(cachedObject.consumers)); consumers == 0 {
		delete(objectStorage.cachedObjects, typeutils.BytesToString(cachedObject.key))
	}
	objectStorage.cacheMutex.Unlock()
}

func (bw *BatchedWriter) runBatchWriter() {

	for atomic.LoadInt32(&bw.running) == 1 {
		bw.writeWg.Add(1)

		if !bw.waitingForTimeout {
			bw.waitingForTimeout = true
			bw.timeoutWg.Add(1)
		}

		wb := bw.badgerInstance.NewWriteBatch()

		writtenValues := make([]*CachedObject, BATCH_WRITER_BATCH_SIZE)
		writtenValuesCounter := 0
	COLLECT_VALUES:
		for writtenValuesCounter < BATCH_WRITER_BATCH_SIZE {
			select {
			case objectToPersist := <-bw.batchQueue:
				bw.writeObject(wb, objectToPersist)
				writtenValues[writtenValuesCounter] = objectToPersist
				writtenValuesCounter++
			case <-time.After(BATCH_WRITER_BATCH_TIMEOUT):
				bw.waitingForTimeout = false
				bw.timeoutWg.Done()

				break COLLECT_VALUES
			}
		}

		if err := wb.Flush(); err != nil && err != badger.ErrBlockedWrites {
			panic(err)
		}

		for _, cachedObject := range writtenValues {
			if cachedObject != nil {
				bw.releaseObject(cachedObject)
			}
		}

		bw.writeWg.Done()
	}
}

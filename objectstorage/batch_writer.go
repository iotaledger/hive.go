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
	badgerInstance *badger.DB
	writeWg        sync.WaitGroup
	startStopMutex syncutils.Mutex
	running        int32
	batchQueue     chan *CachedObjectImpl
}

func NewBatchedWriter(badgerInstance *badger.DB) *BatchedWriter {
	return &BatchedWriter{
		badgerInstance: badgerInstance,
		writeWg:        sync.WaitGroup{},
		startStopMutex: syncutils.Mutex{},
		running:        0,
		batchQueue:     make(chan *CachedObjectImpl, BATCH_WRITER_QUEUE_SIZE),
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

func (bw *BatchedWriter) batchWrite(object *CachedObjectImpl) {
	if atomic.LoadInt32(&bw.running) == 0 {
		bw.StartBatchWriter()
	}

	if atomic.AddInt32(&(object.batchWriteScheduled), 1) == 1 {
		bw.batchQueue <- object
	}
}

func (bw *BatchedWriter) writeObject(writeBatch *badger.WriteBatch, cachedObject *CachedObjectImpl) {
	objectStorage := cachedObject.objectStorage
	if !objectStorage.options.persistenceEnabled {
		if storableObject := cachedObject.Get(); !typeutils.IsInterfaceNil(storableObject) {
			storableObject.SetModified(false)
		}

		return
	}

	if consumers := atomic.LoadInt32(&(cachedObject.consumers)); consumers == 0 {
		if storableObject := cachedObject.Get(); !typeutils.IsInterfaceNil(storableObject) {
			if storableObject.IsDeleted() {
				storableObject.SetModified(false)

				if err := writeBatch.Delete(objectStorage.generatePrefix([][]byte{cachedObject.key})); err != nil {
					panic(err)
				}
			} else if storableObject.PersistenceEnabled() && storableObject.IsModified() {
				storableObject.SetModified(false)

				var marshaledObject []byte
				if !objectStorage.options.keysOnly {
					marshaledObject, _ = storableObject.MarshalBinary()
				}

				if err := writeBatch.Set(objectStorage.generatePrefix([][]byte{cachedObject.key}), marshaledObject); err != nil {
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

func (bw *BatchedWriter) releaseObject(cachedObject *CachedObjectImpl) {
	objectStorage := cachedObject.objectStorage

	objectStorage.cacheMutex.Lock()
	if consumers := atomic.LoadInt32(&(cachedObject.consumers)); consumers == 0 {
		// only delete if the object is still empty, or was not modified since the write (and was not evicted yet)
		if storableObject := cachedObject.Get(); (typeutils.IsInterfaceNil(storableObject) || !storableObject.IsModified()) && atomic.AddInt32(&cachedObject.evicted, 1) == 1 && objectStorage.deleteElementFromCache(cachedObject.key) && objectStorage.size == 0 {
			objectStorage.cachedObjectsEmpty.Done()
		}
	}
	objectStorage.cacheMutex.Unlock()
}

func (bw *BatchedWriter) runBatchWriter() {
	for atomic.LoadInt32(&bw.running) == 1 {
		bw.writeWg.Add(1)

		var wb *badger.WriteBatch
		if bw.badgerInstance != nil {
			wb = bw.badgerInstance.NewWriteBatch()
		}

		writtenValues := make([]*CachedObjectImpl, BATCH_WRITER_BATCH_SIZE)
		writtenValuesCounter := 0
	COLLECT_VALUES:
		for writtenValuesCounter < BATCH_WRITER_BATCH_SIZE {
			select {
			case objectToPersist := <-bw.batchQueue:
				atomic.StoreInt32(&(objectToPersist.batchWriteScheduled), 0)

				bw.writeObject(wb, objectToPersist)
				writtenValues[writtenValuesCounter] = objectToPersist
				writtenValuesCounter++
			case <-time.After(BATCH_WRITER_BATCH_TIMEOUT):
				break COLLECT_VALUES
			}
		}

		if wb != nil {
			if err := wb.Flush(); err != nil && err != badger.ErrBlockedWrites {
				panic(err)
			}
		}

		for _, cachedObject := range writtenValues {
			if cachedObject != nil {
				bw.releaseObject(cachedObject)
			}
		}

		bw.writeWg.Done()
	}
}

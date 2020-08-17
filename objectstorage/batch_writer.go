package objectstorage

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/syncutils"
	"github.com/iotaledger/hive.go/typeutils"
)

const (
	BatchWriterQueueSize    = BatchWriterBatchSize
	BatchWriterBatchSize    = 10000
	BatchWriterBatchTimeout = 500 * time.Millisecond
)

type BatchedWriter struct {
	store          kvstore.KVStore
	writeWg        sync.WaitGroup
	startStopMutex syncutils.Mutex
	autoStartOnce  sync.Once
	running        int32
	scheduledCount int32
	batchQueue     chan *CachedObjectImpl
}

func NewBatchedWriter(store kvstore.KVStore) *BatchedWriter {
	return &BatchedWriter{
		store:          store,
		writeWg:        sync.WaitGroup{},
		startStopMutex: syncutils.Mutex{},
		running:        0,
		batchQueue:     make(chan *CachedObjectImpl, BatchWriterQueueSize),
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
	bw.autoStartOnce.Do(func() {
		if atomic.LoadInt32(&bw.running) == 0 {
			bw.StartBatchWriter()
		}
	})

	// abort if the BatchWriter has been stopped
	if atomic.LoadInt32(&bw.running) == 0 {
		return
	}

	// abort if the very same object has been queued already
	if atomic.AddInt32(&(object.batchWriteScheduled), 1) != 1 {
		return
	}

	// queue object
	atomic.AddInt32(&bw.scheduledCount, 1)
	bw.batchQueue <- object
}

func (bw *BatchedWriter) writeObject(batchedMuts kvstore.BatchedMutations, cachedObject *CachedObjectImpl) {
	objectStorage := cachedObject.objectStorage
	if !objectStorage.options.persistenceEnabled {
		if storableObject := cachedObject.Get(); !typeutils.IsInterfaceNil(storableObject) {
			storableObject.SetModified(false)
		}

		return
	}

	consumers := atomic.LoadInt32(&(cachedObject.consumers))
	if consumers < 0 {
		panic("too many unregistered consumers of cached object")
	}

	storableObject := cachedObject.Get()

	if typeutils.IsInterfaceNil(storableObject) {
		// only blind delete if there are no consumers
		if consumers == 0 && cachedObject.blindDelete.IsSet() {
			if err := batchedMuts.Delete(cachedObject.key); err != nil {
				panic(err)
			}
		}

		return
	}

	if storableObject.IsDeleted() {
		// only delete if there are no consumers
		if consumers == 0 {
			storableObject.SetModified(false)
			if err := batchedMuts.Delete(cachedObject.key); err != nil {
				panic(err)
			}
		}

		return
	}

	// only store if there are no consumers anymore or the object should be stored on creation
	if consumers != 0 && !cachedObject.objectStorage.options.storeOnCreation {
		return
	}

	if storableObject.ShouldPersist() && storableObject.IsModified() {
		storableObject.SetModified(false)

		var marshaledValue []byte
		if !objectStorage.options.keysOnly {
			marshaledValue = storableObject.ObjectStorageValue()
		}

		if err := batchedMuts.Set(cachedObject.key, marshaledValue); err != nil {
			panic(err)
		}
	}
}

func (bw *BatchedWriter) releaseObject(cachedObject *CachedObjectImpl) {
	objectStorage := cachedObject.objectStorage

	objectStorage.flushMutex.RLock()
	defer objectStorage.flushMutex.RUnlock()

	objectStorage.cacheMutex.Lock()
	defer objectStorage.cacheMutex.Unlock()

	if consumers := atomic.LoadInt32(&(cachedObject.consumers)); consumers == 0 {
		// only delete if the object is still empty, or was not modified since the write (and was not evicted yet)
		if storableObject := cachedObject.Get(); (typeutils.IsInterfaceNil(storableObject) || !storableObject.IsModified()) && atomic.AddInt32(&cachedObject.evicted, 1) == 1 && objectStorage.deleteElementFromCache(cachedObject.key) && objectStorage.size == 0 {
			objectStorage.cachedObjectsEmpty.Done()
		}
	}
}

func (bw *BatchedWriter) runBatchWriter() {
	bw.writeWg.Add(1)

	for atomic.LoadInt32(&bw.running) == 1 || atomic.LoadInt32(&bw.scheduledCount) != 0 {
		var batchedMuts kvstore.BatchedMutations

		writtenValues := make([]*CachedObjectImpl, BatchWriterBatchSize)
		writtenValuesCounter := 0
	CollectValues:
		for writtenValuesCounter < BatchWriterBatchSize {
			select {
			case objectToPersist := <-bw.batchQueue:

				if batchedMuts == nil && bw.store != nil {
					batchedMuts = bw.store.Batched()
				}

				atomic.StoreInt32(&(objectToPersist.batchWriteScheduled), 0)
				atomic.AddInt32(&bw.scheduledCount, -1)

				bw.writeObject(batchedMuts, objectToPersist)
				writtenValues[writtenValuesCounter] = objectToPersist
				writtenValuesCounter++
			case <-time.After(BatchWriterBatchTimeout):
				break CollectValues
			}
		}

		if batchedMuts != nil {
			if err := batchedMuts.Commit(); err != nil {
				panic(err)
			}
		}

		// release written values
		for i := 0; i < writtenValuesCounter; i++ {
			bw.releaseObject(writtenValues[i])
		}
	}

	bw.writeWg.Done()
}

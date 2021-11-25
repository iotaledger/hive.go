package kvstore

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/iotaledger/hive.go/v2/syncutils"
)

// BatchWriteObject is an object that can be persisted to the KVStore in batches using the BatchedWriter.
type BatchWriteObject interface {
	// BatchWrite mashalls the object and adds it to the BatchedMutations.
	BatchWrite(batchedMuts BatchedMutations)
	// BatchWriteDone is called after the object was persisted.
	BatchWriteDone()
	// BatchWriteScheduled returns true if the object is already scheduled for a BatchWrite operation.
	BatchWriteScheduled() bool
	// ResetBatchWriteScheduled resets the flag that the object is scheduled for a BatchWrite operation.
	ResetBatchWriteScheduled()
}

const (
	BatchWriterQueueSize    = BatchWriterBatchSize
	BatchWriterBatchSize    = 10000
	BatchWriterBatchTimeout = 500 * time.Millisecond
)

// BatchedWriter persists BatchWriteObjects in batches to a KVStore.
type BatchedWriter struct {
	store          KVStore
	writeWg        sync.WaitGroup
	startStopMutex syncutils.Mutex
	autoStartOnce  sync.Once
	running        int32
	scheduledCount int32
	batchQueue     chan BatchWriteObject
}

// NewBatchedWriter creates a new BatchedWriter instance.
func NewBatchedWriter(store KVStore) *BatchedWriter {
	return &BatchedWriter{
		store:          store,
		writeWg:        sync.WaitGroup{},
		startStopMutex: syncutils.Mutex{},
		running:        0,
		batchQueue:     make(chan BatchWriteObject, BatchWriterQueueSize),
	}
}

// KVStore returns the underlying KVStore.
func (bw *BatchedWriter) KVStore() KVStore {
	return bw.store
}

// startBatchWriter starts the batch writer if it was not started yet.
func (bw *BatchedWriter) startBatchWriter() {
	bw.startStopMutex.Lock()
	if atomic.LoadInt32(&bw.running) == 0 {
		atomic.StoreInt32(&bw.running, 1)
		go bw.runBatchWriter()
	}
	bw.startStopMutex.Unlock()
}

// StopBatchWriter stops the batch writer and waits until all enqueued objects are written.
func (bw *BatchedWriter) StopBatchWriter() {
	bw.startStopMutex.Lock()
	if atomic.LoadInt32(&bw.running) != 0 {
		atomic.StoreInt32(&bw.running, 0)

		bw.writeWg.Wait()
	}
	bw.startStopMutex.Unlock()
}

// Enqueue adds a BatchWriteObject to the write queue.
// It also starts the batch writer if not done yet.
func (bw *BatchedWriter) Enqueue(object BatchWriteObject) {
	bw.autoStartOnce.Do(func() {
		if atomic.LoadInt32(&bw.running) == 0 {
			bw.startBatchWriter()
		}
	})

	// abort if the BatchWriter has been stopped
	if atomic.LoadInt32(&bw.running) == 0 {
		return
	}

	// abort if the very same object has been queued already
	if object.BatchWriteScheduled() {
		return
	}

	// queue object
	atomic.AddInt32(&bw.scheduledCount, 1)
	bw.batchQueue <- object
}

// runBatchWriter collects objects in batches and persists them to the KVStore.
func (bw *BatchedWriter) runBatchWriter() {
	bw.writeWg.Add(1)

	for atomic.LoadInt32(&bw.running) == 1 || atomic.LoadInt32(&bw.scheduledCount) != 0 {
		var batchedMuts BatchedMutations

		writtenValues := make([]BatchWriteObject, BatchWriterBatchSize)
		batchWriterTimeoutChan := time.After(BatchWriterBatchTimeout)
		writtenValuesCounter := 0
	CollectValues:
		for writtenValuesCounter < BatchWriterBatchSize {
			select {
			case objectToPersist := <-bw.batchQueue:

				if batchedMuts == nil && bw.store != nil {
					batchedMuts = bw.store.Batched()
				}

				objectToPersist.ResetBatchWriteScheduled()
				atomic.AddInt32(&bw.scheduledCount, -1)

				objectToPersist.BatchWrite(batchedMuts)
				writtenValues[writtenValuesCounter] = objectToPersist
				writtenValuesCounter++

			case <-batchWriterTimeoutChan:
				break CollectValues
			}
		}

		if batchedMuts != nil {
			if err := batchedMuts.Commit(); err != nil {
				panic(err)
			}
		}

		for i := 0; i < writtenValuesCounter; i++ {
			writtenValues[i].BatchWriteDone()
		}
	}

	bw.writeWg.Done()
}

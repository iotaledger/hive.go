package kvstore

import (
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/iotaledger/hive.go/syncutils"
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
	running        *atomic.Bool
	scheduledCount *atomic.Int32
	batchQueue     chan BatchWriteObject
}

// NewBatchedWriter creates a new BatchedWriter instance.
func NewBatchedWriter(store KVStore) *BatchedWriter {
	return &BatchedWriter{
		store:          store,
		writeWg:        sync.WaitGroup{},
		startStopMutex: syncutils.Mutex{},
		running:        atomic.NewBool(false),
		scheduledCount: atomic.NewInt32(0),
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
	if !bw.running.Load() {
		bw.running.Store(true)
		go bw.runBatchWriter()
	}
	bw.startStopMutex.Unlock()
}

// StopBatchWriter stops the batch writer and waits until all enqueued objects are written.
func (bw *BatchedWriter) StopBatchWriter() {
	bw.startStopMutex.Lock()
	if bw.running.Load() {
		bw.running.Store(false)

		bw.writeWg.Wait()
	}
	bw.startStopMutex.Unlock()
}

// Enqueue adds a BatchWriteObject to the write queue.
// It also starts the batch writer if not done yet.
func (bw *BatchedWriter) Enqueue(object BatchWriteObject) {
	bw.autoStartOnce.Do(func() {
		if !bw.running.Load() {
			bw.startBatchWriter()
		}
	})

	// abort if the BatchWriter has been stopped
	if !bw.running.Load() {
		return
	}

	// abort if the very same object has been queued already
	if object.BatchWriteScheduled() {
		return
	}

	// queue object
	bw.scheduledCount.Inc()
	bw.batchQueue <- object
}

// runBatchWriter collects objects in batches and persists them to the KVStore.
func (bw *BatchedWriter) runBatchWriter() {
	bw.writeWg.Add(1)

	for bw.running.Load() || bw.scheduledCount.Load() != 0 {
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
				bw.scheduledCount.Dec()

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

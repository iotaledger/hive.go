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
	flushChan      chan struct{}
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
		flushChan:      make(chan struct{}),
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

// Flush sends a signal to flush all the queued elements.
func (bw *BatchedWriter) Flush() {
	if bw.running.Load() {
		select {
		case bw.flushChan <- struct{}{}:
		default:
			// another flush request is already queued => no need to block
		}
	}
}

// runBatchWriter collects objects in batches and persists them to the KVStore.
func (bw *BatchedWriter) runBatchWriter() {
	bw.writeWg.Add(1)

	for bw.running.Load() || bw.scheduledCount.Load() != 0 {

		batchCollector := newBatchCollector(bw.store.Batched(), bw.scheduledCount, BatchWriterBatchSize)
		batchWriterTimeoutChan := time.After(BatchWriterBatchTimeout)
		shouldFlush := false

	CollectValues:
		for {
			select {

			// an element was added to the queue
			case objectToPersist := <-bw.batchQueue:
				if batchCollector.Add(objectToPersist) {
					// batch size was reached => apply the mutations
					if err := batchCollector.Commit(); err != nil {
						panic(err)
					}
					break CollectValues
				}

			// flush was triggered
			case <-bw.flushChan:
				shouldFlush = true
				break CollectValues

			// batch timeout was reached
			case <-batchWriterTimeoutChan:
				// apply the collected mutations
				if err := batchCollector.Commit(); err != nil {
					panic(err)
				}
				break CollectValues
			}
		}

		if shouldFlush {
			// flush was triggered, collect all remaining elements from the queue and commit them.

		FlushValues:
			for {
				select {

				// pick the next element from the queue
				case objectToPersist := <-bw.batchQueue:
					if batchCollector.Add(objectToPersist) {
						// batch size was reached => apply the mutations
						if err := batchCollector.Commit(); err != nil {
							panic(err)
						}
						// create a new collector to batch the remaining elements
						batchCollector = newBatchCollector(bw.store.Batched(), bw.scheduledCount, BatchWriterBatchSize)
					}

				// no elements left
				default:
					// apply the collected mutations
					if err := batchCollector.Commit(); err != nil {
						panic(err)
					}
					break FlushValues
				}
			}
		}
	}

	bw.writeWg.Done()
}

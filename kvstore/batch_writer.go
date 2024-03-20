package kvstore

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/iotaledger/hive.go/runtime/syncutils"
	"github.com/iotaledger/hive.go/runtime/timeutil"
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

// the default options applied to the BatchedWriter.
var defaultOptions = []Option{
	WithQueueSize(10000),
	WithBatchSize(10000),
	WithBatchTimeout(500 * time.Millisecond),
}

// Options define options for the BatchedWriter.
type Options struct {
	// the size of the batch queue.
	queueSize int
	// the maximum amount of elements in the batch.
	batchSize int
	// the timeout for collecting elements for the batch.
	batchTimeout time.Duration
}

// applies the given Option.
func (so *Options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(so)
	}
}

// WithQueueSize defines the size of the batch queue.
func WithQueueSize(queueSize int) Option {
	return func(opts *Options) {
		opts.queueSize = queueSize
	}
}

// WithBatchSize defines the maximum amount of elements in the batch.
func WithBatchSize(batchSize int) Option {
	return func(opts *Options) {
		opts.batchSize = batchSize
	}
}

// WithBatchTimeout defines the timeout for collecting elements for the batch.
func WithBatchTimeout(batchTimeout time.Duration) Option {
	return func(opts *Options) {
		opts.batchTimeout = batchTimeout
	}
}

// Option is a function setting a BatchedWriter option.
type Option func(opts *Options)

// BatchedWriter persists BatchWriteObjects in batches to a KVStore.
type BatchedWriter struct {
	store          KVStore
	writeWg        sync.WaitGroup
	startStopMutex syncutils.Mutex
	autoStartOnce  sync.Once
	running        atomic.Bool
	scheduledCount atomic.Int32
	batchQueue     chan BatchWriteObject
	flushChan      chan struct{}
	opts           *Options
}

// NewBatchedWriter creates a new BatchedWriter instance.
func NewBatchedWriter(store KVStore, opts ...Option) *BatchedWriter {
	options := &Options{}
	options.apply(defaultOptions...)
	options.apply(opts...)

	return &BatchedWriter{
		store:          store,
		writeWg:        sync.WaitGroup{},
		startStopMutex: syncutils.Mutex{},
		batchQueue:     make(chan BatchWriteObject, options.queueSize),
		flushChan:      make(chan struct{}, 1), // must be buffered with size 1 since no receiver is actively waiting.
		opts:           options,
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
	bw.scheduledCount.Add(1)
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
		batchedMutation, err := bw.store.Batched()
		if err != nil {
			panic(err)
		}
		batchCollector := newBatchCollector(batchedMutation, &bw.scheduledCount, bw.opts.batchSize)
		shouldFlush := false

		collectValues := func() {
			batchWriterTimeoutTimer := time.NewTimer(bw.opts.batchTimeout)
			defer timeutil.CleanupTimer(batchWriterTimeoutTimer)

			for {
				select {
				// an element was added to the queue
				case objectToPersist := <-bw.batchQueue:
					if batchCollector.Add(objectToPersist) {
						// batch size was reached => apply the mutations
						if err := batchCollector.Commit(); err != nil {
							panic(err)
						}

						return
					}

				// flush was triggered
				case <-bw.flushChan:
					shouldFlush = true

					return

				// batch timeout was reached
				case <-batchWriterTimeoutTimer.C:
					// apply the collected mutations
					if err := batchCollector.Commit(); err != nil {
						panic(err)
					}

					return
				}
			}
		}

		collectValues()

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
						batchedMutation, err := bw.store.Batched()
						if err != nil {
							panic(err)
						}
						batchCollector = newBatchCollector(batchedMutation, &bw.scheduledCount, bw.opts.batchSize)
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

package kvstore

import (
	"sync/atomic"
)

// BatchCollector is used to collect objects that should be written.
type BatchCollector struct {
	batchedMuts          BatchedMutations
	scheduledCount       *atomic.Int32
	batchSize            int
	writtenValues        []BatchWriteObject
	writtenValuesCounter int
	committed            bool
}

// newBatchCollector creates a new BatchCollector that is used to collect objects that should be written.
func newBatchCollector(batchedMuts BatchedMutations, scheduledCount *atomic.Int32, batchSize int) *BatchCollector {
	return &BatchCollector{
		batchedMuts:          batchedMuts,
		scheduledCount:       scheduledCount,
		batchSize:            batchSize,
		writtenValues:        make([]BatchWriteObject, batchSize),
		writtenValuesCounter: 0,
		committed:            false,
	}
}

// Add adds an object to the batch.
// It returns true in case the batch size is reached.
func (br *BatchCollector) Add(objectToPersist BatchWriteObject) (batchSizeReached bool) {
	if br.committed {
		panic("mutations were already committed")
	}

	objectToPersist.ResetBatchWriteScheduled()
	br.scheduledCount.Add(-1)

	objectToPersist.BatchWrite(br.batchedMuts)
	br.writtenValues[br.writtenValuesCounter] = objectToPersist
	br.writtenValuesCounter++

	return br.writtenValuesCounter >= br.batchSize
}

// Commit applies the collected mutations.
func (br *BatchCollector) Commit() error {
	if br.committed {
		panic("mutations were already committed")
	}
	br.committed = true

	if br.writtenValuesCounter == 0 {
		// nothing to commit
		br.batchedMuts.Cancel()

		return nil
	}

	if err := br.batchedMuts.Commit(); err != nil {
		return err
	}

	for i := range br.writtenValuesCounter {
		br.writtenValues[i].BatchWriteDone()
	}

	return nil
}

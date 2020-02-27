package objectstorage

import (
	"sync"
)

type RetainedPartition struct {
	childPartitions map[string]*RetainedPartition
	retainCounter   int
	mutex           sync.RWMutex
}

func NewRetainedPartition() *RetainedPartition {
	return &RetainedPartition{
		childPartitions: make(map[string]*RetainedPartition),
	}
}

func (retainedPartition *RetainedPartition) IsEmpty() bool {
	retainedPartition.mutex.RLock()
	defer retainedPartition.mutex.RUnlock()

	return retainedPartition.isEmpty()
}

func (retainedPartition *RetainedPartition) isEmpty() bool {
	return retainedPartition.retainCounter == 0 && len(retainedPartition.childPartitions) == 0
}

func (retainedPartition *RetainedPartition) IsRetained(keys []string) bool {
	retainedPartition.mutex.RLock()
	defer retainedPartition.mutex.RUnlock()

	if len(keys) == 0 {
		return !retainedPartition.isEmpty()
	}

	childPartition, childPartitionExists := retainedPartition.childPartitions[keys[0]]
	if !childPartitionExists {
		return false
	}

	return childPartition.IsRetained(keys[1:])
}

func (retainedPartition *RetainedPartition) Release(keysToRelease []string) bool {
	if len(keysToRelease) == 0 {
		retainedPartition.mutex.Lock()
		retainedPartition.retainCounter--
		retainedPartition.mutex.Unlock()
	} else {
		childPartitionKey := keysToRelease[0]

		retainedPartition.mutex.RLock()
		childPartition := retainedPartition.childPartitions[childPartitionKey]
		retainedPartition.mutex.RUnlock()

		if childPartition.Release(keysToRelease[1:]) {
			retainedPartition.mutex.Lock()
			if childPartition.IsEmpty() {
				delete(retainedPartition.childPartitions, childPartitionKey)
			}
			retainedPartition.mutex.Unlock()
		}
	}

	return retainedPartition.isEmpty()
}

func (retainedPartition *RetainedPartition) Retain(keysToRetain []string) {
	if len(keysToRetain) == 0 {
		retainedPartition.mutex.Lock()
		retainedPartition.retainCounter++
		retainedPartition.mutex.Unlock()

		return
	}

	partitionKey := keysToRetain[0]

	retainedPartition.mutex.RLock()
	if childPartition, childPartitionExists := retainedPartition.childPartitions[partitionKey]; childPartitionExists {
		childPartition.Retain(keysToRetain[1:])

		retainedPartition.mutex.RUnlock()

		return
	}
	retainedPartition.mutex.RUnlock()

	retainedPartition.mutex.Lock()
	defer retainedPartition.mutex.Unlock()

	if childPartition, childPartitionExists := retainedPartition.childPartitions[partitionKey]; childPartitionExists {
		childPartition.Retain(keysToRetain[1:])

		return
	}

	newChildPartition := NewRetainedPartition()
	retainedPartition.childPartitions[partitionKey] = newChildPartition
	newChildPartition.Retain(keysToRetain[1:])
}

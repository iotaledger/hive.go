package objectstorage

import (
	"sync"
)

type PartitionsManager struct {
	childPartitions map[string]*PartitionsManager
	retainCounter   int
	mutex           sync.RWMutex
}

func NewPartitionsManager() *PartitionsManager {
	return &PartitionsManager{
		childPartitions: make(map[string]*PartitionsManager),
	}
}

func (partitionsManager *PartitionsManager) IsEmpty() bool {
	partitionsManager.mutex.RLock()
	defer partitionsManager.mutex.RUnlock()

	return partitionsManager.isEmpty()
}

func (partitionsManager *PartitionsManager) isEmpty() bool {
	return partitionsManager.retainCounter == 0 && len(partitionsManager.childPartitions) == 0
}

func (partitionsManager *PartitionsManager) IsRetained(keys []string) bool {
	partitionsManager.mutex.RLock()
	defer partitionsManager.mutex.RUnlock()

	if len(keys) == 0 {
		return !partitionsManager.isEmpty()
	}

	childPartition, childPartitionExists := partitionsManager.childPartitions[keys[0]]
	if !childPartitionExists {
		return false
	}

	return childPartition.IsRetained(keys[1:])
}

func (partitionsManager *PartitionsManager) Release(keysToRelease []string) bool {
	if len(keysToRelease) == 0 {
		partitionsManager.mutex.Lock()
		partitionsManager.retainCounter--
		partitionsManager.mutex.Unlock()
	} else {
		childPartitionKey := keysToRelease[0]

		partitionsManager.mutex.RLock()
		childPartition := partitionsManager.childPartitions[childPartitionKey]
		partitionsManager.mutex.RUnlock()

		if childPartition.Release(keysToRelease[1:]) {
			partitionsManager.mutex.Lock()
			if childPartition.IsEmpty() {
				delete(partitionsManager.childPartitions, childPartitionKey)
			}
			partitionsManager.mutex.Unlock()
		}
	}

	return partitionsManager.isEmpty()
}

func (partitionsManager *PartitionsManager) Retain(keysToRetain []string) {
	if len(keysToRetain) == 0 {
		partitionsManager.mutex.Lock()
		partitionsManager.retainCounter++
		partitionsManager.mutex.Unlock()

		return
	}

	partitionKey := keysToRetain[0]

	partitionsManager.mutex.RLock()
	if childPartition, childPartitionExists := partitionsManager.childPartitions[partitionKey]; childPartitionExists {
		childPartition.Retain(keysToRetain[1:])

		partitionsManager.mutex.RUnlock()

		return
	}
	partitionsManager.mutex.RUnlock()

	partitionsManager.mutex.Lock()
	defer partitionsManager.mutex.Unlock()

	if childPartition, childPartitionExists := partitionsManager.childPartitions[partitionKey]; childPartitionExists {
		childPartition.Retain(keysToRetain[1:])

		return
	}

	newChildPartition := NewPartitionsManager()
	partitionsManager.childPartitions[partitionKey] = newChildPartition
	newChildPartition.Retain(keysToRetain[1:])
}

// FreeMemory copies the content of the internal maps to newly created maps.
// This is necessary, otherwise the GC is not able to free the memory used by the old maps.
// "delete" doesn't shrink the maximum memory used by the map, since it only marks the entry as deleted.
func (partitionsManager *PartitionsManager) FreeMemory() {
	partitionsManager.mutex.Lock()
	defer partitionsManager.mutex.Unlock()

	childPartitions := make(map[string]*PartitionsManager)
	for key, childPartition := range partitionsManager.childPartitions {
		childPartitions[key] = childPartition
		if childPartition != nil {
			childPartition.FreeMemory()
		}
	}
	partitionsManager.childPartitions = childPartitions
}

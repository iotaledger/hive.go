package syncutils

import (
	"sync"
)

// region MultiMutex ///////////////////////////////////////////////////////////////////////////////////////////////////

type MultiMutex struct {
	locks     map[interface{}]empty
	locksCond *sync.Cond
	initOnce  sync.Once
}

func NewMultiMutex() *MultiMutex {
	return &MultiMutex{
		locks: make(map[interface{}]empty),
		locksCond: &sync.Cond{
			L: &sync.Mutex{},
		},
	}
}

// LockEntity locks all locks that are required for the given LockEntity.
func (mutex *MultiMutex) LockEntity(entity LockableEntity) {
	mutex.Lock(entity.Locks()...)
}

// UnlockEntity unlocks all locks that are required for the given LockEntity.
func (mutex *MultiMutex) UnlockEntity(entity LockableEntity) {
	mutex.Unlock(entity.Locks()...)
}

func (mutex *MultiMutex) Lock(identifiers ...interface{}) {
	mutex.initOnce.Do(func() {
		mutex.locks = make(map[interface{}]empty)
		mutex.locksCond = &sync.Cond{L: &sync.Mutex{}}
	})

	mutex.locksCond.L.Lock()

AcquireLocks:
	for i, identifier := range identifiers {
		if _, isLocked := mutex.locks[identifier]; !isLocked {
			mutex.locks[identifier] = void
		} else {
			for j := 0; j < i; j++ {
				delete(mutex.locks, identifiers[j])
			}

			if i > 0 {
				mutex.locksCond.Broadcast()
			}
			mutex.locksCond.Wait()

			goto AcquireLocks
		}
	}

	mutex.locksCond.L.Unlock()
}

func (mutex *MultiMutex) Unlock(identifiers ...interface{}) {
	mutex.initOnce.Do(func() {
		mutex.locks = make(map[interface{}]empty)
		mutex.locksCond = &sync.Cond{L: &sync.Mutex{}}
	})

	mutex.locksCond.L.Lock()
	for _, identifier := range identifiers {
		delete(mutex.locks, identifier)
	}
	mutex.locksCond.L.Unlock()
	mutex.locksCond.Broadcast()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region LockableEntity ///////////////////////////////////////////////////////////////////////////////////////////////

// LockableEntity is an interface that allows to lock (and unlock) entities that are generating complex locks.
type LockableEntity interface {
	// Locks returns the locks that the entity needs to lock.
	Locks() (locks []interface{})
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

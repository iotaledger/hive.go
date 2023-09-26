package syncutils

import (
	"sync"

	"github.com/izuc/zipp.foundation/ds/types"
)

// region MultiMutex ///////////////////////////////////////////////////////////////////////////////////////////////////

// MultiMutex is a mutex that allows to lock multiple entities exclusively.
// Entities are represented by interface{} identifiers and can be presented in arbitrary order.
// Goroutine 1        Goroutine 2        Goroutine 3
//
//	Lock(a,c)           		-          Lock(c,b) <- blocking
//	     work            Lock(b)   	        wait
//
// Unlock(a,c)               work   	        wait
//   - Unlock(b)   	        wait
//   - -   	   Lock(c,b) <- successful
type MultiMutex struct {
	locks     map[interface{}]types.Empty
	locksCond *sync.Cond
}

// NewMultiMutex creates a new MultiMutex.
func NewMultiMutex() *MultiMutex {
	return &MultiMutex{
		locks:     make(map[interface{}]types.Empty),
		locksCond: sync.NewCond(&sync.Mutex{}),
	}
}

// LockEntity locks all locks that are required for the given LockableEntity.
func (m *MultiMutex) LockEntity(entity LockableEntity) {
	m.Lock(entity.Locks()...)
}

// UnlockEntity unlocks all locks that are required for the given LockableEntity.
func (m *MultiMutex) UnlockEntity(entity LockableEntity) {
	m.Unlock(entity.Locks()...)
}

// Lock blocks until all locks given by ids can be acquired atomically.
func (m *MultiMutex) Lock(ids ...interface{}) {
	m.locksCond.L.Lock()

TryAcquireLocks:
	for _, id := range ids {
		if _, exists := m.locks[id]; exists {
			m.locksCond.Wait()

			goto TryAcquireLocks
		}
	}

	// Acquire all locks.
	for _, id := range ids {
		m.locks[id] = types.Void
	}

	m.locksCond.L.Unlock()
}

// Unlock releases the locks of ids.
func (m *MultiMutex) Unlock(ids ...interface{}) {
	m.locksCond.L.Lock()
	for _, id := range ids {
		delete(m.locks, id)
	}
	m.locksCond.L.Unlock()
	m.locksCond.Broadcast()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region LockableEntity ///////////////////////////////////////////////////////////////////////////////////////////////

// LockableEntity is an interface that allows to lock (and unlock) entities that are generating complex locks.
type LockableEntity interface {
	// Locks returns the locks that the entity needs to lock.
	Locks() (locks []interface{})
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

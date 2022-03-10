package syncutils

import (
	"sync"

	"github.com/iotaledger/hive.go/datastructure/set"
)

// region MultiMutex ///////////////////////////////////////////////////////////////////////////////////////////////////

// MultiMutex is a mutex that allows to lock multiple entities exclusively.
// Entities are represented by interface{} identifiers and can be presented in arbitrary order.
// Goroutine 1        Goroutine 2        Goroutine 3
//   Lock(a,c)           		-          Lock(c,b) <- blocking
//        work            Lock(b)   	        wait
// Unlock(a,c)               work   	        wait
//           -          Unlock(b)   	        wait
//           -                  -   	   Lock(c,b) <- successful
type MultiMutex struct {
	locks     set.Set
	locksCond *sync.Cond
}

// NewMultiMutex creates a new MultiMutex.
func NewMultiMutex() *MultiMutex {
	return &MultiMutex{
		locks:     set.New(),
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

func (m *MultiMutex) allLocksFree(ids []interface{}) bool {
	for _, id := range ids {
		if m.locks.Has(id) {
			return false
		}
	}

	return true
}

func (m *MultiMutex) Lock(ids ...interface{}) {
	m.locksCond.L.Lock()

	for !m.allLocksFree(ids) {
		m.locksCond.Wait()
	}

	// Acquire all locks.
	for _, id := range ids {
		m.locks.Add(id)
	}

	// AcquireLocks:
	// 	for i, identifier := range ids {
	// 		// Loop and function are only exited if all locks could be acquired.
	// 		if m.locks.Add(identifier) {
	// 			continue
	// 		}
	//
	// 		// Remove previously, optimistically added locks because they are not actually locked.
	// 		// We need to wait and repeat the whole process again until we can acquire all locks at once.
	// 		for j := 0; j < i; j++ {
	// 			m.locks.Delete(ids[j])
	// 		}
	//
	// 		// TODO: Why wake up others? We should only be waiting on locks that are currently in use.
	// 		// As soon as something gets Unlocked we wake up and try to acquire all locks again.
	// 		// This process might actually create lock contention on locksCond.L: whenever we unlock e.g. a message, we will check
	// 		// for all currently waiting routines whether we can now lock all of them.
	// 		// Maybe we can introduce a cond for each key so that we only wake up relevant routines.
	// 		if i > 0 {
	// 			m.locksCond.Broadcast()
	// 		}
	// 		m.locksCond.Wait()
	//
	// 		goto AcquireLocks
	// 	}

	m.locksCond.L.Unlock()
}

func (m *MultiMutex) Unlock(identifiers ...interface{}) {
	m.locksCond.L.Lock()
	for _, identifier := range identifiers {
		m.locks.Delete(identifier)
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

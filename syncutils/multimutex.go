package syncutils

import (
	"sync"
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
	m     *sync.Mutex
	locks map[interface{}]*condLock
}

type condLock struct {
	writerCond *sync.Cond
	// readerCond *sync.Cond
	//
	// CurrentReaders uint32
	// CurrentWriter  bool
	// WaitingWriters uint32
}

func newCondLock(m *sync.Mutex) *condLock {
	return &condLock{
		writerCond: sync.NewCond(m),
	}
}

// NewMultiMutex creates a new MultiMutex.
func NewMultiMutex() *MultiMutex {
	return &MultiMutex{
		locks: make(map[interface{}]*condLock),
		m:     &sync.Mutex{},
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

func (m *MultiMutex) allLocksFree(ids []interface{}) (allLockFree bool, firstLockedCond *condLock) {
	for _, id := range ids {
		if lock, exists := m.locks[id]; exists {
			return false, lock
		}
	}

	return true, nil
}

func (m *MultiMutex) Lock(ids ...interface{}) {
	m.m.Lock()

TryAcquireLocks:
	for _, id := range ids {
		if lock, exists := m.locks[id]; exists {
			lock.writerCond.Wait()
			goto TryAcquireLocks
		}
	}

	// for free, lock := m.allLocksFree(ids); !free; free, lock = m.allLocksFree(ids) {
	// 	lock.writerCond.Wait()
	// }

	// Acquire all locks.
	for _, id := range ids {
		m.locks[id] = newCondLock(m.m)
	}

	m.m.Unlock()
}

func (m *MultiMutex) Unlock(ids ...interface{}) {
	broadcastLocks := make([]*condLock, 0, len(ids))
	m.m.Lock()
	for _, id := range ids {
		broadcastLocks = append(broadcastLocks, m.locks[id])
		delete(m.locks, id)
	}
	m.m.Unlock()

	for _, l := range broadcastLocks {
		l.writerCond.Broadcast()
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region LockableEntity ///////////////////////////////////////////////////////////////////////////////////////////////

// LockableEntity is an interface that allows to lock (and unlock) entities that are generating complex locks.
type LockableEntity interface {
	// Locks returns the locks that the entity needs to lock.
	Locks() (locks []interface{})
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

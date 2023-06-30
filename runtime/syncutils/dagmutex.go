package syncutils

import (
	"sync"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/ierrors"
)

// DAGMutex is a multi-entity reader/writer mutual exclusion lock that allows for starvation.
// Entities can be registered dynamically by providing a comparable identifier.
// The structure and relation of these entities MUST NOT contain any cycles as this may lead to deadlocks while waiting
// to acquire locks cyclically.
//
// Entities can be Lock-ed one at a time. The call blocks until the lock for this entity can be acquired.
// Entities can be RLock-ed multiple in arbitrary order. The call blocks until all read locks can be acquired.
//
// Consider the following example of a DAG with 3 entities A,B,C.
// ┌───┬─┐
// │   │B◄─┐
// ┌▼┐ └─┘ │
// │A│     │
// └▲┘    ┌┴┐
// │      │C│
// └──────┴─┘
//
// Let's assume the following 3 goroutines are competing for access. Time is advancing with every row.
// Goroutine 1        Goroutine 2        Goroutine 3
//
//	  Lock(A)           		-         RLock(A,B) <- blocking, not able to acquire locks
//	     work            Lock(B)   	        wait
//	Unlock(A)               work   	        wait
//	        -          Unlock(B)   	        wait <- (internally) now RLock(A) is successful, but still waiting for B
//	        -                  -   	 RLock(A, B) <- successful acquired, holding the locks now
type DAGMutex[T comparable] struct {
	consumerCounter *shrinkingmap.ShrinkingMap[T, int]
	mutexes         *shrinkingmap.ShrinkingMap[T, *StarvingMutex]
	sync.Mutex
}

// NewDAGMutex creates a new DAGMutex.
func NewDAGMutex[T comparable]() *DAGMutex[T] {
	return &DAGMutex[T]{
		consumerCounter: shrinkingmap.New[T, int](),
		mutexes:         shrinkingmap.New[T, *StarvingMutex](),
	}
}

// RLock locks all given entities for reading.
// It blocks until all read locks can be acquired.
//
// It should not be used for recursive read locking.
// A blocked Lock call DOES NOT exclude new readers from acquiring the lock. Hence, it is starving.
func (d *DAGMutex[T]) RLock(ids ...T) {
	for _, mutex := range d.registerMutexes(ids...) {
		mutex.RLock()
	}
}

// RUnlock unlocks reading for all given entities.
// It does not affect other simultaneous readers.
func (d *DAGMutex[T]) RUnlock(ids ...T) {
	for _, mutex := range d.unregisterMutexes(ids...) {
		mutex.RUnlock()
	}
}

// Lock locks the given entity for writing.
// If the lock is already locked for reading or writing, Lock blocks until the lock is available.
func (d *DAGMutex[T]) Lock(id T) {
	d.Mutex.Lock()
	mutex := d.registerMutex(id)
	d.Mutex.Unlock()

	mutex.Lock()
}

// Unlock unlocks the given entity for writing.
//
// As with Mutexes, a locked DAGMutex is not associated with a particular goroutine. One goroutine may RLock (Lock) an
// entity within DAGMutex and then arrange for another goroutine to RUnlock (Unlock) it.
func (d *DAGMutex[T]) Unlock(id T) {
	d.Mutex.Lock()
	mutex := d.unregisterMutex(id)
	if mutex == nil {
		d.Mutex.Unlock()

		return
	}
	d.Mutex.Unlock()

	mutex.Unlock()
}

func (d *DAGMutex[T]) registerMutexes(ids ...T) (mutexes []*StarvingMutex) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	mutexes = make([]*StarvingMutex, len(ids))
	for i, id := range ids {
		mutexes[i] = d.registerMutex(id)
	}

	return mutexes
}

func (d *DAGMutex[T]) registerMutex(id T) (mutex *StarvingMutex) {
	mutex, mutexExists := d.mutexes.Get(id)
	if !mutexExists {
		mutex = NewStarvingMutex()
		d.mutexes.Set(id, mutex)
	}
	count, _ := d.consumerCounter.Get(id)
	d.consumerCounter.Set(id, count+1)

	return mutex
}

func (d *DAGMutex[T]) unregisterMutexes(ids ...T) (mutexes []*StarvingMutex) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	mutexes = make([]*StarvingMutex, 0)
	for _, id := range ids {
		if mutex := d.unregisterMutex(id); mutex != nil {
			mutexes = append(mutexes, mutex)
		}
	}

	return mutexes
}

func (d *DAGMutex[T]) unregisterMutex(id T) (mutex *StarvingMutex) {
	if count, _ := d.consumerCounter.Get(id); count == 1 {
		d.consumerCounter.Delete(id)
		d.mutexes.Delete(id)

		return nil
	}

	mutex, mutexExists := d.mutexes.Get(id)
	if !mutexExists {
		panic(ierrors.Errorf("called Unlock or RUnlock too often for entity with %v", id))
	}
	count, _ := d.consumerCounter.Get(id)
	d.consumerCounter.Set(id, count-1)

	return mutex
}

package syncutils

import (
	"fmt"
	"sync"
)

type DAGMutex[T comparable] struct {
	consumerCounter map[T]int
	mutexes         map[T]*StarvingMutex
	sync.Mutex
}

func NewDAGMutex[T comparable]() *DAGMutex[T] {
	return &DAGMutex[T]{
		consumerCounter: make(map[T]int),
		mutexes:         make(map[T]*StarvingMutex),
	}
}

func (d *DAGMutex[T]) RLock(ids ...T) {
	for _, mutex := range d.registerMutexes(ids...) {
		mutex.RLock()
	}
}

func (d *DAGMutex[T]) RUnlock(ids ...T) {
	for _, mutex := range d.unregisterMutexes(ids...) {
		mutex.RUnlock()
	}
}

func (d *DAGMutex[T]) Lock(id T) {
	d.Mutex.Lock()
	mutex := d.registerMutex(id)
	d.Mutex.Unlock()

	mutex.Lock()
}

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
	mutex, mutexExists := d.mutexes[id]
	if !mutexExists {
		mutex = NewStarvingMutex()
		d.mutexes[id] = mutex
	}

	d.consumerCounter[id]++

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
	if d.consumerCounter[id] == 1 {
		delete(d.consumerCounter, id)
		delete(d.mutexes, id)
		return nil
	}

	mutex, mutexExists := d.mutexes[id]
	if !mutexExists {
		panic(fmt.Errorf("called Unlock or RUnlock too often for entity with %v", id))
	}

	d.consumerCounter[id]--

	return mutex
}

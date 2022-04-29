package syncutils

import (
	"fmt"
	"sync"

	"github.com/iotaledger/hive.go/debug"
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
	for i, mutex := range d.registerMutexes(ids...) {
		if debug.Enabled {
			fmt.Println(debug.GoroutineID(), "RLock", ids[i])
		}
		mutex.RLock()
		if debug.Enabled {
			fmt.Println(debug.GoroutineID(), "RLocked", ids[i])
		}
	}
}

func (d *DAGMutex[T]) RUnlock(ids ...T) {
	for i, mutex := range d.unregisterMutexes(ids...) {
		if debug.Enabled {
			fmt.Println(debug.GoroutineID(), "RUnLock", ids[i])
		}
		mutex.RUnlock()
		if debug.Enabled {
			fmt.Println(debug.GoroutineID(), "RUnLocked", ids[i])
		}
	}
}

func (d *DAGMutex[T]) Lock(id T) {
	d.Mutex.Lock()
	mutex := d.registerMutex(id)
	d.Mutex.Unlock()

	if debug.Enabled {
		fmt.Println(debug.GoroutineID(), "Lock", id)
	}
	mutex.Lock()
	if debug.Enabled {
		fmt.Println(debug.GoroutineID(), "Locked", id)
	}
}

func (d *DAGMutex[T]) Unlock(id T) {
	d.Mutex.Lock()
	mutex := d.unregisterMutex(id)
	if mutex == nil {
		d.Mutex.Unlock()
		if debug.Enabled {
			fmt.Println(debug.GoroutineID(), "UnLock (early)", id)
		}
		if debug.Enabled {
			fmt.Println(debug.GoroutineID(), "UnLocked (early)", id)
		}
		return
	}
	d.Mutex.Unlock()

	if debug.Enabled {
		fmt.Println(debug.GoroutineID(), "UnLock", id)
	}
	mutex.Unlock()
	if debug.Enabled {
		fmt.Println(debug.GoroutineID(), "UnLocked", id)
	}
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

		// we don't need to unlock removed mutexes as nobody else is using them anymore anyway
		return nil
	}

	mutex, mutexExists := d.mutexes[id]
	if !mutexExists {
		panic(fmt.Errorf("called Unlock or RUnlock too often for entity with %v", id))
	}

	d.consumerCounter[id]--

	return mutex
}

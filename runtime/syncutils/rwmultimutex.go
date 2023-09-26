package syncutils

import (
	"sync"
)

type RWMultiMutex struct {
	rLocks    map[interface{}]int
	wLocks    map[interface{}]empty
	locksCond *sync.Cond
	initOnce  sync.Once
}

func (mutex *RWMultiMutex) Lock(identifiers ...interface{}) {
	mutex.initOnce.Do(func() {
		mutex.rLocks = make(map[interface{}]int)
		mutex.wLocks = make(map[interface{}]empty)
		mutex.locksCond = &sync.Cond{L: &sync.Mutex{}}
	})

	mutex.locksCond.L.Lock()

AcquireLocks:
	for i, identifier := range identifiers {
		if _, isWLocked := mutex.wLocks[identifier]; isWLocked {
			for j := 0; j < i; j++ {
				delete(mutex.wLocks, identifiers[j])
			}

			if i > 0 {
				mutex.locksCond.Broadcast()
			}
			mutex.locksCond.Wait()

			goto AcquireLocks
		} else {
			if _, isRLocked := mutex.rLocks[identifier]; isRLocked {
				for j := 0; j < i; j++ {
					delete(mutex.wLocks, identifiers[j])
				}

				if i > 0 {
					mutex.locksCond.Broadcast()
				}
				mutex.locksCond.Wait()

				goto AcquireLocks
			} else {
				mutex.wLocks[identifier] = void
			}
		}
	}

	mutex.locksCond.L.Unlock()
}

func (mutex *RWMultiMutex) RLock(identifiers ...interface{}) {
	mutex.initOnce.Do(func() {
		mutex.rLocks = make(map[interface{}]int)
		mutex.wLocks = make(map[interface{}]empty)
		mutex.locksCond = &sync.Cond{L: &sync.Mutex{}}
	})

	mutex.locksCond.L.Lock()

AcquireLocks:
	for i, identifier := range identifiers {
		if _, isWLocked := mutex.wLocks[identifier]; isWLocked {
			for j := 0; j < i; j++ {
				if currentReaderCount := mutex.rLocks[identifiers[j]]; currentReaderCount == 1 {
					delete(mutex.rLocks, identifiers[j])
				} else {
					mutex.rLocks[identifiers[j]] = currentReaderCount - 1
				}
			}

			if i > 0 {
				mutex.locksCond.Broadcast()
			}
			mutex.locksCond.Wait()

			goto AcquireLocks
		} else {
			mutex.rLocks[identifier]++
		}
	}

	mutex.locksCond.L.Unlock()
}

func (mutex *RWMultiMutex) RUnlock(identifiers ...interface{}) {
	mutex.initOnce.Do(func() {
		mutex.rLocks = make(map[interface{}]int)
		mutex.wLocks = make(map[interface{}]empty)
		mutex.locksCond = &sync.Cond{L: &sync.Mutex{}}
	})

	mutex.locksCond.L.Lock()

	for _, identifier := range identifiers {
		if currentReaderCount := mutex.rLocks[identifier]; currentReaderCount == 1 {
			delete(mutex.rLocks, identifier)
		} else {
			mutex.rLocks[identifier] = currentReaderCount - 1
		}
	}

	mutex.locksCond.L.Unlock()
	mutex.locksCond.Broadcast()
}

func (mutex *RWMultiMutex) Unlock(identifiers ...interface{}) {
	mutex.initOnce.Do(func() {
		mutex.rLocks = make(map[interface{}]int)
		mutex.wLocks = make(map[interface{}]empty)
		mutex.locksCond = &sync.Cond{L: &sync.Mutex{}}
	})

	mutex.locksCond.L.Lock()

	for _, identifier := range identifiers {
		delete(mutex.wLocks, identifier)
	}

	mutex.locksCond.L.Unlock()
	mutex.locksCond.Broadcast()
}

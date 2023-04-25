package syncutils

import (
	"sync"
)

type RWMutexFake struct {
	sync.RWMutex
}

func (m *RWMutexFake) RLock() {
	m.Lock()
}

func (m *RWMutexFake) RUnlock() {
	m.Unlock()
}

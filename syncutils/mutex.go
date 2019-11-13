// +build !deadlock

package syncutils

import (
	"sync"
)

type Mutex sync.Mutex
type RWMutex sync.RWMutex

func (m *Mutex) Lock()   { (*sync.Mutex)(m).Lock() }
func (m *Mutex) Unlock() { (*sync.Mutex)(m).Unlock() }

func (m *RWMutex) Lock()    { (*sync.RWMutex)(m).Lock() }
func (m *RWMutex) Unlock()  { (*sync.RWMutex)(m).Unlock() }
func (m *RWMutex) RLock()   { (*sync.RWMutex)(m).RLock() }
func (m *RWMutex) RUnlock() { (*sync.RWMutex)(m).RUnlock() }

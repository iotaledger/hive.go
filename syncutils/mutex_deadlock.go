// +build deadlock

package syncutils

import (
	deadlock "github.com/sasha-s/go-deadlock"
	"time"
)

type Mutex deadlock.Mutex
type RWMutex deadlock.RWMutex

func init() {
	deadlock.Opts.DeadlockTimeout = time.Duration(20 * time.Second)
}

func (m *Mutex) Lock()   { (*deadlock.Mutex)(m).Lock() }
func (m *Mutex) Unlock() { (*deadlock.Mutex)(m).Unlock() }

func (m *RWMutex) Lock()    { (*deadlock.RWMutex)(m).Lock() }
func (m *RWMutex) Unlock()  { (*deadlock.RWMutex)(m).Unlock() }
func (m *RWMutex) RLock()   { (*deadlock.RWMutex)(m).RLock() }
func (m *RWMutex) RUnlock() { (*deadlock.RWMutex)(m).RUnlock() }

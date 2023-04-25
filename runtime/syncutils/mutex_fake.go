//go:build fakemutex

package syncutils

type Mutex = RWMutexFake
type RWMutex = RWMutexFake

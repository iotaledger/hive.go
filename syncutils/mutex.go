// +build !deadlock

package syncutils

import (
	"sync"
)

type Mutex = sync.Mutex
type RWMutex = sync.RWMutex

// +build deadlock

package syncutils

import (
	deadlock "github.com/sasha-s/go-deadlock"
	"time"
)

type Mutex = deadlock.Mutex
type RWMutex = deadlock.RWMutex

func init() {
	deadlock.Opts.DeadlockTimeout = time.Duration(20 * time.Second)
}

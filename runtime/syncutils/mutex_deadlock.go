//go:build deadlock
// +build deadlock

package syncutils

import (
	"time"

	"github.com/sasha-s/go-deadlock"
)

type Mutex = deadlock.Mutex
type RWMutex = deadlock.RWMutex

func init() {
	deadlock.Opts.DeadlockTimeout = time.Duration(20 * time.Second)
}

//go:build deadlock
// +build deadlock

package syncutils

import (
	"fmt"
	"time"

	"github.com/sasha-s/go-deadlock"
)

type Mutex = deadlock.Mutex
type RWMutex = deadlock.RWMutex

func init() {
	fmt.Println(">>>> use deadlock mutex")
	deadlock.Opts.DeadlockTimeout = time.Duration(20 * time.Second)
}

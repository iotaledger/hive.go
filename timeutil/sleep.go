package timeutil

import (
	"github.com/iotaledger/hive.go/daemon"
	"time"
)

func Sleep(interval time.Duration) bool {
	select {
	case <-daemon.ShutdownSignal:
		return false

	case <-time.After(interval):
		return true
	}
}

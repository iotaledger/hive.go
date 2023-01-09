package event

import (
	"github.com/iotaledger/hive.go/core/workerpool"
)

var Loop *workerpool.UnboundedWorkerPool

func init() {
	Loop = workerpool.NewUnboundedWorkerPool()
	Loop.Start()
}

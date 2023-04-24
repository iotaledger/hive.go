//go:build fake
// +build fake

package syncutils

import (
	"github.com/iotaledger/hive.go/runtime/syncutils"
)

type Mutex = syncutils.RWMutexFake
type RWMutex = syncutils.RWMutexFake

func init() {
	fmt.Println(">>>> use fake mutex")
}

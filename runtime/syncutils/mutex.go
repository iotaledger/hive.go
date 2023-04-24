//go:build !deadlock && !fake

package syncutils

import (
	"fmt"
	"sync"
)

type Mutex = sync.Mutex
type RWMutex = sync.RWMutex

func init() {
	fmt.Println(">>>> use fake mutex")
}

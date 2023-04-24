//go:build fake
// +build fake

package syncutils

import "fmt"

type Mutex = RWMutexFake
type RWMutex = RWMutexFake

func init() {
	fmt.Println(">>>> use fake mutex")
}

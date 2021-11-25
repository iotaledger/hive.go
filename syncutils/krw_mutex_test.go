package syncutils_test

import (
	"github.com/iotaledger/hive.go/syncutils"
	"testing"
)

func TestKRWMutex_Free(t *testing.T) {
	krwMutex := syncutils.NewKRWMutex()

	krwMutex.Register("test")
	krwMutex.Register("test")
	krwMutex.Free("test")
	krwMutex.Free("test")
}

func BenchmarkKRWMutex(b *testing.B) {
	krwMutex := syncutils.NewKRWMutex()

	for i := 0; i < b.N; i++ {
		krwMutex.Register(i)
		krwMutex.Free(i)
	}
}

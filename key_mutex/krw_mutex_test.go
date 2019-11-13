package key_mutex_test

import (
	key_mutex "github.com/iotaledger/hive.go/key_mutex"
	"testing"
)

func TestKRWMutex_Free(t *testing.T) {
	krwMutex := key_mutex.NewKRWMutex()

	krwMutex.Register("test")
	krwMutex.Register("test")
	krwMutex.Free("test")
	krwMutex.Free("test")
}

func BenchmarkKRWMutex(b *testing.B) {
	krwMutex := key_mutex.NewKRWMutex()

	for i := 0; i < b.N; i++ {
		krwMutex.Register(i)
		krwMutex.Free(i)
	}
}

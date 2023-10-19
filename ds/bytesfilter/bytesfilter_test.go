package bytesfilter_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ds/bytesfilter"
)

func newIdentifier(bytes []byte) [32]byte {
	return blake2b.Sum256(bytes)
}

func TestBytesFilter(t *testing.T) {

	filter := bytesfilter.New[[32]byte](newIdentifier, 2)

	data := randBytes(20)
	id, added := filter.Add(data)
	require.True(t, added)

	exists := filter.Contains(data)
	require.True(t, exists)
	exists = filter.ContainsIdentifier(id)
	require.True(t, exists)

	// add new identifier
	randData := rand32ByteArray()
	randID := newIdentifier(randData[:])
	added = filter.AddIdentifier(randID)
	require.True(t, added)
	exists = filter.ContainsIdentifier(randID)
	require.True(t, exists)
	exists = filter.Contains(randData[:])
	require.True(t, exists)

	// add existing identifier
	added = filter.AddIdentifier(randID)
	require.False(t, added)

	// add existing bytes
	id1, added := filter.Add(data)
	require.False(t, added)
	require.ElementsMatch(t, id, id1)

	tmpID := rand32ByteArray()
	exists = filter.ContainsIdentifier(newIdentifier(tmpID[:]))
	require.False(t, exists)

	data3 := randBytes(20)
	id3, added := filter.Add(data3)
	require.True(t, added)
	exists = filter.Contains(data3)
	require.True(t, exists)
	exists = filter.ContainsIdentifier(id3)
	require.True(t, exists)

	// the first element should be removed
	exists = filter.Contains(data)
	require.False(t, exists)
	exists = filter.ContainsIdentifier(id)
	require.False(t, exists)
}

func BenchmarkAdd(b *testing.B) {
	filter, bytesFilter := setupTest(15000, 1604)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		filter.Add(bytesFilter)
	}
}

func BenchmarkContains(b *testing.B) {
	filter, bytesFilter := setupTest(15000, 1604)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		filter.Contains(bytesFilter)
	}
}

func setupTest(filterSize int, byteArraySize int) (*bytesfilter.BytesFilter[[32]byte], []byte) {
	filter := bytesfilter.New[[32]byte](newIdentifier, filterSize)

	for j := 0; j < filterSize; j++ {
		byteArray := make([]byte, byteArraySize)

		for i := 0; i < len(byteArray); i++ {
			byteArray[(i+j)%byteArraySize] = byte((i + j) % 128)
		}

		filter.Add(byteArray)
	}

	byteArray := make([]byte, byteArraySize)

	for i := 0; i < len(byteArray); i++ {
		byteArray[i] = byte(i % 128)
	}

	return filter, byteArray
}

func randBytes(length int) []byte {
	var b []byte
	for i := 0; i < length; i++ {
		b = append(b, byte(rand.Intn(127)))
	}
	return b
}

func rand32ByteArray() [32]byte {
	var h [32]byte
	b := randBytes(32)
	copy(h[:], b)
	return h
}

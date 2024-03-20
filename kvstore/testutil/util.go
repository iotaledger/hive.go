package testutil

import (
	"bytes"
	"math/rand"
	"sort"
	"sync"
	"time"
)

var (
	//nolint:gosec // we do not care about weak random numbers here
	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	randLock   = &sync.Mutex{}
)

func RandomRead(p []byte) (n int, err error) {
	// Rand needs to be locked: https://github.com/golang/go/issues/3611
	randLock.Lock()
	defer randLock.Unlock()

	return seededRand.Read(p)
}

func RandomIntn(n int) int {
	// Rand needs to be locked: https://github.com/golang/go/issues/3611
	randLock.Lock()
	defer randLock.Unlock()

	return seededRand.Intn(n)
}

func RandomInt31n(n int32) int32 {
	// Rand needs to be locked: https://github.com/golang/go/issues/3611
	randLock.Lock()
	defer randLock.Unlock()

	return seededRand.Int31n(n)
}

func RandomInt63n(n int64) int64 {
	// Rand needs to be locked: https://github.com/golang/go/issues/3611
	randLock.Lock()
	defer randLock.Unlock()

	return seededRand.Int63n(n)
}

func RandomFloat64() float64 {
	// Rand needs to be locked: https://github.com/golang/go/issues/3611
	randLock.Lock()
	defer randLock.Unlock()

	return seededRand.Float64()
}

// RandByte returns a random byte.
func RandByte() byte {
	return byte(RandomIntn(256))
}

// RandBytes returns length amount random bytes.
func RandBytes(length int) []byte {
	b := make([]byte, 0, length)
	for range length {
		b = append(b, byte(RandomIntn(256)))
	}

	return b
}

func RandString(length int) string {
	return string(RandBytes(length))
}

// RandUint8 returns a random uint8.
func RandUint8(max uint8) uint8 {
	return uint8(RandomInt31n(int32(max)))
}

// RandUint16 returns a random uint16.
func RandUint16(max uint16) uint16 {
	return uint16(RandomInt31n(int32(max)))
}

// RandUint32 returns a random uint32.
func RandUint32(max uint32) uint32 {
	return uint32(RandomInt63n(int64(max)))
}

// RandUint64 returns a random uint64.
func RandUint64(max uint64) uint64 {
	return uint64(RandomInt63n(int64(uint32(max))))
}

// RandFloat64 returns a random float64.
func RandFloat64(max float64) float64 {
	return RandomFloat64() * max
}

// Rand32ByteArray returns an array with 32 random bytes.
func Rand32ByteArray() [32]byte {
	var h [32]byte
	b := RandBytes(32)
	copy(h[:], b)

	return h
}

// Rand49ByteArray returns an array with 49 random bytes.
func Rand49ByteArray() [49]byte {
	var h [49]byte
	b := RandBytes(49)
	copy(h[:], b)

	return h
}

// Rand64ByteArray returns an array with 64 random bytes.
func Rand64ByteArray() [64]byte {
	var h [64]byte
	b := RandBytes(64)
	copy(h[:], b)

	return h
}

// SortedRand32BytArray returns a count length slice of sorted 32 byte arrays.
func SortedRand32BytArray(count int) [][32]byte {
	hashes := make([][32]byte, count)
	for i := range count {
		hashes[i] = Rand32ByteArray()
	}
	sort.Slice(hashes, func(i, j int) bool {
		return bytes.Compare(hashes[i][:], hashes[j][:]) < 0
	})

	return hashes
}

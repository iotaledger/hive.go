package crypto

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
)

// RandomnessSource implements a cryptographically secure randomness source.
type RandomnessSource struct{}

// Seed is usually used to initialize the generator to a deterministic state but we want real randomness.
func (s RandomnessSource) Seed(int64) {}

// Int63 returns a non-negative random 63-bit integer as an int64.
func (s RandomnessSource) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

// Uint64 returns a random 64-bit value as a uint64.
func (s RandomnessSource) Uint64() (v uint64) {
	if err := binary.Read(crand.Reader, binary.BigEndian, &v); err != nil {
		panic(err)
	}

	return v
}

// Randomness contains an Rand instance of the previously defined RandomnessSource.
//
//nolint:gosec // false positive
var Randomness = rand.New(RandomnessSource{})

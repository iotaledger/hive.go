package bls

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"

	"go.dedis.ch/kyber/v3/util/random"
)

// secureRandomnessSource implements a cryptographically secure randomness source.
type secureRandomnessSource struct{}

// Seed is usually used to initialize the generator to a deterministic state but we want real randomness.
func (s secureRandomnessSource) Seed(int64) {}

// Int63 returns a non-negative random 63-bit integer as an int64.
func (s secureRandomnessSource) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

// Uint64 returns a random 64-bit value as a uint64.
func (s secureRandomnessSource) Uint64() (v uint64) {
	if err := binary.Read(crand.Reader, binary.BigEndian, &v); err != nil {
		panic(err)
	}

	return v
}

// randomness contains an instance of the previously defined secureRandomnessSource that is used by BLS.
var randomness = random.New(rand.New(secureRandomnessSource{}))

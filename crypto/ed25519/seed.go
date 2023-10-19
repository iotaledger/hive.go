package ed25519

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/subtle"
	"encoding/binary"

	"github.com/mr-tron/base58"
	"golang.org/x/crypto/blake2b"
)

// Seed is a generator for a deterministic sequence of KeyPairs.
type Seed struct {
	seedBytes []byte
}

// NewSeed represents the factory method for a Seed object. It either generates a new random seed or restores an
// existing one from a sequence of bytes.
func NewSeed(optionalSeedBytes ...[]byte) *Seed {
	if len(optionalSeedBytes) >= 1 {
		if len(optionalSeedBytes[0]) < SeedSize {
			panic("seed is not long enough")
		}

		return &Seed{
			seedBytes: optionalSeedBytes[0],
		}
	}

	randomSeedBytes := make([]byte, ed25519.SeedSize)
	_, err := rand.Read(randomSeedBytes)
	if err != nil {
		panic(err)
	}

	return &Seed{
		seedBytes: randomSeedBytes,
	}
}

// KeyPair retrieves the n'th KeyPair from the Seed.
func (seed *Seed) KeyPair(n uint64) (keyPair *KeyPair) {
	keyPair = &KeyPair{}

	privateKey := ed25519.NewKeyFromSeed(seed.subSeed(n))
	//nolint:forcetypeassert // false positive, we know it's an ed25519.PublicKey
	publicKey := privateKey.Public().(ed25519.PublicKey)

	copy(keyPair.PrivateKey[:], privateKey)
	copy(keyPair.PublicKey[:], publicKey)

	return
}

// Bytes marshals the Seed object into a sequence of Bytes that can be used to later restore the Seed by handing it into
// the factory method.
func (seed *Seed) Bytes() []byte {
	return seed.seedBytes
}

// subSeed generates the n'th sub seed of this Seed which is then used to generate the KeyPair.
func (seed *Seed) subSeed(n uint64) (subSeed []byte) {
	subSeed = make([]byte, SeedSize)

	indexBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(indexBytes, n)
	hashOfIndexBytes := blake2b.Sum256(indexBytes)

	subtle.XORBytes(subSeed, seed.seedBytes, hashOfIndexBytes[:])

	return
}

// String returns a human-readable version of the Seed (base58 encoded).
func (seed *Seed) String() string {
	return base58.Encode(seed.seedBytes)
}

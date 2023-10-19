package ed25519

import (
	"crypto/ed25519"

	"github.com/mr-tron/base58"
)

// PrivateKey is the type of Ed25519 private keys.
type PrivateKey [PrivateKeySize]byte

// PrivateKeyFromBytes creates a PrivateKey from the given bytes.
func PrivateKeyFromBytes(bytes []byte) (result PrivateKey, consumedBytes int, err error) {
	if len(bytes) < PrivateKeySize {
		return PrivateKey{}, 0, ErrNotEnoughBytes
	}

	return PrivateKey(bytes), PrivateKeySize, nil
}

// PrivateKeyFromSeed calculates a private key from a seed.
func PrivateKeyFromSeed(seed []byte) (result PrivateKey) {
	copy(result[:], ed25519.NewKeyFromSeed(seed))

	return
}

// Sign signs the message with privateKey and returns a signature.
func (privateKey PrivateKey) Sign(data []byte) (result Signature) {
	copy(result[:], ed25519.Sign(privateKey[:], data))

	return
}

// Public returns the PublicKey corresponding to privateKey.
func (privateKey PrivateKey) Public() (result PublicKey) {
	//nolint:forcetypeassert // false positive, we know it's an ed25519.PublicKey
	publicKey := ed25519.PrivateKey(privateKey[:]).Public().(ed25519.PublicKey)
	copy(result[:], publicKey)

	return
}

// Bytes returns the privateKey in bytes.
func (privateKey PrivateKey) Bytes() ([]byte, error) {
	return privateKey[:], nil
}

// String returns a human-readable version of the PrivateKey (base58 encoded).
func (privateKey PrivateKey) String() string {
	return base58.Encode(privateKey[:])
}

// Seed returns the private key seed corresponding to privateKey. It is provided for
// interoperability with RFC 8032. RFC 8032's private keys correspond to seeds
// in this package.
func (privateKey PrivateKey) Seed() *Seed {
	bytes := ed25519.PrivateKey(privateKey[:]).Seed()
	seed := NewSeed(bytes)

	return seed
}

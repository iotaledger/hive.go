package ed25519

import (
	"github.com/mr-tron/base58"
)

type Signature [SignatureSize]byte

// SignatureFromBytes creates a Signature from the given bytes.
func SignatureFromBytes(bytes []byte) (result Signature, consumedBytes int, err error) {
	if len(bytes) < SignatureSize {
		return EmptySignature, 0, ErrNotEnoughBytes
	}

	return Signature(bytes), SignatureSize, nil
}

// Bytes returns the signature in bytes.
func (signature Signature) Bytes() ([]byte, error) {
	return signature[:], nil
}

// String returns a human-readable version of the Signature (base58 encoded).
func (signature Signature) String() string {
	return base58.Encode(signature[:])
}

var EmptySignature Signature

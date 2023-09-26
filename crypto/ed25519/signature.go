package ed25519

import (
	"github.com/mr-tron/base58"

	"github.com/izuc/zipp.foundation/serializer/v2/marshalutil"
)

type Signature [SignatureSize]byte

// SignatureFromBytes creates a Signature from the given bytes.
func SignatureFromBytes(bytes []byte) (result Signature, consumedBytes int, err error) {
	consumedBytes, err = (&result).FromBytes(bytes)
	return
}

func ParseSignature(marshalUtil *marshalutil.MarshalUtil) (Signature, error) {
	id, err := marshalUtil.Parse(func(data []byte) (interface{}, int, error) { return SignatureFromBytes(data) })
	if err != nil {
		return Signature{}, err
	}

	return id.(Signature), nil
}

// FromBytes initializes Signature from the given bytes.
func (signature *Signature) FromBytes(bytes []byte) (consumedBytes int, err error) {
	if len(bytes) < SignatureSize {
		return 0, ErrNotEnoughBytes
	}

	copy(signature[:SignatureSize], bytes)

	return SignatureSize, nil
}

// Bytes returns the signature in bytes.
func (signature Signature) Bytes() ([]byte, error) {
	return signature[:], nil
}

// String returns a human-readable version of the Signature (base58 encoded).
func (signature Signature) String() string {
	return base58.Encode(signature[:])
}

func (signature *Signature) UnmarshalBinary(bytes []byte) error {
	if _, err := signature.FromBytes(bytes); err != nil {
		return err
	}

	return nil
}

var EmptySignature Signature

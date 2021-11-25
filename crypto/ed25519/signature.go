package ed25519

import (
	"errors"
	"fmt"

	"github.com/mr-tron/base58"

	"github.com/iotaledger/hive.go/marshalutil"
)

type Signature [SignatureSize]byte

// SignatureFromBytes creates a Signature from the given bytes.
func SignatureFromBytes(bytes []byte) (result Signature, consumedBytes int, err error) {
	if len(bytes) < SignatureSize {
		err = fmt.Errorf("bytes too short")
		return
	}

	copy(result[:SignatureSize], bytes)
	consumedBytes = SignatureSize

	return
}

func ParseSignature(marshalUtil *marshalutil.MarshalUtil) (Signature, error) {
	if id, err := marshalUtil.Parse(func(data []byte) (interface{}, int, error) { return SignatureFromBytes(data) }); err != nil {
		return Signature{}, err
	} else {
		return id.(Signature), nil
	}
}

// Bytes returns the signature in bytes.
func (signature Signature) Bytes() []byte {
	return signature[:]
}

// String returns a human readable version of the Signature (base58 encoded).
func (signature Signature) String() string {
	return base58.Encode(signature[:])
}

func (signature *Signature) UnmarshalBinary(bytes []byte) (err error) {
	if len(bytes) < SignatureSize {
		return errors.New("not enough bytes")
	}

	copy(signature[:], bytes)
	return
}

var EmptySignature Signature

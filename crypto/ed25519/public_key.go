package ed25519

import (
	"errors"
	"fmt"

	"github.com/oasislabs/ed25519"

	"github.com/mr-tron/base58"

	"github.com/iotaledger/hive.go/marshalutil"
)

// PublicKey is the type of Ed25519 public keys.
type PublicKey [PublicKeySize]byte

// PublicKeyFromBytes creates a PublicKey from the given bytes.
func PublicKeyFromBytes(bytes []byte) (result PublicKey, err error, consumedBytes int) {
	if len(bytes) < PublicKeySize {
		err = fmt.Errorf("bytes too short")
		return
	}

	copy(result[:], bytes)
	consumedBytes = PublicKeySize

	return
}

// RecoverKey makes sure that key and signature have the correct length
// and verifies whether sig is a valid signature of data by pub.
func RecoverKey(key, data, sig []byte) (result PublicKey, err error) {
	if l := len(key); l != PublicKeySize {
		err = fmt.Errorf("invalid key length: %d, need %d", l, PublicKeySize)
		return
	}
	if l := len(sig); l != SignatureSize {
		err = fmt.Errorf("invalid signature length: %d, need %d", l, SignatureSize)
		return
	}
	if !ed25519.Verify(key, data, sig) {
		err = fmt.Errorf("invalid signature")
		return
	}

	copy(result[:], key)
	return
}

func ParsePublicKey(marshalUtil *marshalutil.MarshalUtil) (PublicKey, error) {
	if id, err := marshalUtil.Parse(func(data []byte) (interface{}, error, int) { return PublicKeyFromBytes(data) }); err != nil {
		return PublicKey{}, err
	} else {
		return id.(PublicKey), nil
	}
}

// VerifySignature reports whether signature is a valid signature of message by publicKey.
func (publicKey PublicKey) VerifySignature(data []byte, signature Signature) bool {
	return ed25519.Verify(publicKey[:], data, signature[:])
}

// Bytes returns the publicKey in bytes.
func (publicKey PublicKey) Bytes() []byte {
	return publicKey[:]
}

// String returns a human readable version of the PublicKey (base58 encoded).
func (publicKey PublicKey) String() string {
	return base58.Encode(publicKey[:])
}

func (publicKey *PublicKey) UnmarshalBinary(bytes []byte) (err error) {
	if len(bytes) < PublicKeySize {
		return errors.New("not enough bytes")
	}

	copy(publicKey[:], bytes[:])

	return
}

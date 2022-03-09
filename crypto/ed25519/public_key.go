package ed25519

import (
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/xerrors"

	"github.com/mr-tron/base58"
	"github.com/oasisprotocol/ed25519"

	"github.com/iotaledger/hive.go/marshalutil"
)

// PublicKey is the type of Ed25519 public keys.
type PublicKey [PublicKeySize]byte

// PublicKeyFromString parses the given string with base58 encoding and returns a PublicKey.
func PublicKeyFromString(s string) (publicKey PublicKey, err error) {
	b, err := base58.Decode(s)
	if err != nil {
		return publicKey, xerrors.Errorf("failed to parse public key %s from base58 string: %w", s, err)
	}
	publicKey, _, err = PublicKeyFromBytes(b)
	return publicKey, err
}

// PublicKeyFromBytes creates a PublicKey from the given bytes.
func PublicKeyFromBytes(bytes []byte) (result PublicKey, consumedBytes int, err error) {
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
	if id, err := marshalUtil.Parse(func(data []byte) (interface{}, int, error) { return PublicKeyFromBytes(data) }); err != nil {
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

// MarshalJSON serializes public key to JSON as base58 encoded string.
func (publicKey PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(publicKey.String())
}

// UnmarshalJSON parses public key from JSON in base58 encoding.
func (publicKey *PublicKey) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	pk, err := PublicKeyFromString(s)
	if err != nil {
		return fmt.Errorf("failed to parse public key from JSON: %w", err)
	}
	*publicKey = pk
	return nil
}

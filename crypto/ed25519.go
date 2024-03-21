package crypto

import (
	"crypto/ed25519"
	"encoding/hex"

	"github.com/iotaledger/hive.go/ierrors"
)

var (
	ErrInvalidKeyLength = ierrors.New("invalid key length")
)

// ParseEd25519PublicKeyFromString parses an ed25519 public key from a string.
func ParseEd25519PublicKeyFromString(key string) (ed25519.PublicKey, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}

	if len(keyBytes) != ed25519.PublicKeySize {
		return nil, ErrInvalidKeyLength
	}

	return ed25519.PublicKey(keyBytes), nil
}

// ParseEd25519PrivateKeyFromString parses an ed25519 private key from a string.
func ParseEd25519PrivateKeyFromString(key string) (ed25519.PrivateKey, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}

	if len(keyBytes) != ed25519.PrivateKeySize {
		return nil, ErrInvalidKeyLength
	}

	return ed25519.PrivateKey(keyBytes), nil
}

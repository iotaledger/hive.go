package ed25519

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
)

const (
	PublicKeySize  = ed25519.PublicKeySize
	SignatureSize  = ed25519.SignatureSize
	PrivateKeySize = ed25519.PrivateKeySize
	SeedSize       = ed25519.SeedSize
)

var (
	ErrInvalidSignature = errors.New("invalid crypto")
)

type PrivateKey ed25519.PrivateKey

// Public returns the PublicKey corresponding to priv.
func (priv PrivateKey) Public() PublicKey {
	publicKey := ed25519.PrivateKey(priv).Public()
	return PublicKey(publicKey.(ed25519.PublicKey))
}

func Sign(privateKey PrivateKey, message []byte) []byte {
	return ed25519.Sign(ed25519.PrivateKey(privateKey), message)
}

func NewKeyFromSeed(seed []byte) PrivateKey {
	return PrivateKey(ed25519.NewKeyFromSeed(seed))
}

// generatePrivateKey generates a private key that can be used for Local.
func GeneratePrivateKey() (PrivateKey, error) {
	_, priv, err := ed25519.GenerateKey(nil)
	return PrivateKey(priv), err
}

type PublicKey ed25519.PublicKey

func Verify(key, data, sig []byte) bool {
	return ed25519.Verify(key, data, sig)
}

func RecoverKey(key, data, sig []byte) (PublicKey, error) {
	if l := len(key); l != PublicKeySize {
		return nil, fmt.Errorf("%w: invalid key length: %d, need %d", ErrInvalidSignature, l, PublicKeySize)
	}
	if l := len(sig); l != SignatureSize {
		return nil, fmt.Errorf("%w: invalid crypto length: %d, need %d", ErrInvalidSignature, l, SignatureSize)
	}
	if !Verify(key, data, sig) {
		return nil, ErrInvalidSignature
	}

	publicKey := make([]byte, PublicKeySize)
	copy(publicKey, key)
	return publicKey, nil
}

func GenerateKey(rand io.Reader) (PublicKey, PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	return PublicKey(pub), PrivateKey(priv), err
}

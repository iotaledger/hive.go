package ed25519

import (
	"github.com/oasisprotocol/ed25519"
)

const (
	PublicKeySize  = ed25519.PublicKeySize
	SignatureSize  = ed25519.SignatureSize
	PrivateKeySize = ed25519.PrivateKeySize
	SeedSize       = ed25519.SeedSize
)

// GenerateKey creates a public/private key pair.
func GenerateKey() (publicKey PublicKey, privateKey PrivateKey, err error) {
	pub, priv, genErr := ed25519.GenerateKey(nil)
	copy(publicKey[:], pub)
	copy(privateKey[:], priv)
	err = genErr

	return
}

// GenerateKey creates a private key.
func GeneratePrivateKey() (privateKey PrivateKey, err error) {
	_, privateKey, err = GenerateKey()
	return
}

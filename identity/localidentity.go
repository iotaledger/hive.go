package identity

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
)

// Identity is a local node's identity.
type LocalIdentity struct {
	*Identity
	privateKey ed25519.PrivateKey
}

// NewLocalIdentity creates a new LocalIdentity.
func NewLocalIdentity(publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) *LocalIdentity {
	return &LocalIdentity{
		Identity:   NewIdentity(publicKey),
		privateKey: privateKey,
	}
}

// NewLocalIdentityWithIdentity creates a new LocalIdentity with a given Identity.
func NewLocalIdentityWithIdentity(identity *Identity, privateKey ed25519.PrivateKey) *LocalIdentity {
	return &LocalIdentity{
		Identity:   identity,
		privateKey: privateKey,
	}
}

// Sign signs the message with the local identity's private key and returns a signature.
func (l LocalIdentity) Sign(message []byte) []byte {
	return ed25519.Sign(l.privateKey, message)
}

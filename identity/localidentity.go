package identity

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
)

type LocalIdentity struct {
	*Identity
	privateKey ed25519.PrivateKey
}

func NewLocalIdentity(publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) *LocalIdentity {
	return &LocalIdentity{
		Identity:   NewIdentity(publicKey),
		privateKey: privateKey,
	}
}

func NewLocalIdentityWithIdentity(identity *Identity, privateKey ed25519.PrivateKey) *LocalIdentity {
	return &LocalIdentity{
		Identity:   identity,
		privateKey: privateKey,
	}
}

// Sign signs the message with the identity's private key and returns a crypto.
func (l LocalIdentity) Sign(message []byte) []byte {
	return ed25519.Sign(l.privateKey, message)
}

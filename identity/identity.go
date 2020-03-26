// Package identity implements a node's identity, consisting out of id and public key.
// A LocalIdentity additionally contains the private key and enables signing messages.
package identity

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
)

// LocalIdentity is a node's identity.
type Identity struct {
	id        ID                // comparable node identifier
	publicKey ed25519.PublicKey // public key used to verify signatures
}

// NewIdentity creates a new identity.
func NewIdentity(publicKey ed25519.PublicKey) *Identity {
	return &Identity{
		id:        NewID(publicKey),
		publicKey: publicKey,
	}
}

func (i Identity) ID() ID {
	return i.id
}

func (i Identity) PublicKey() ed25519.PublicKey {
	return i.publicKey
}

func GenerateIdentity() *Identity {
	publicKey, _, err := ed25519.GenerateKey()
	if err != nil {
		panic(err)
	}

	return &Identity{
		id:        NewID(publicKey),
		publicKey: publicKey,
	}
}

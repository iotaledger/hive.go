package identity

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
)

type Identity struct {
	id        ID                // comparable node identifier
	publicKey ed25519.PublicKey // public key used to verify signatures
}

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

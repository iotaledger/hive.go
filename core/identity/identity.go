// Package identity implements a node's identity, consisting out of id and public key.
// A LocalIdentity additionally contains the private key and enables signing messages.
package identity

import (
	ed255192 "github.com/iotaledger/hive.go/core/crypto/ed25519"
	"github.com/iotaledger/hive.go/core/marshalutil"
)

// LocalIdentity is a node's identity.
type Identity struct {
	id        ID                 // comparable node identifier
	publicKey ed255192.PublicKey // public key used to verify signatures
}

// New creates a new identity from the given PublicKey.
func New(publicKey ed255192.PublicKey) *Identity {
	return &Identity{
		id:        NewID(publicKey),
		publicKey: publicKey,
	}
}

func Parse(marshalUtil *marshalutil.MarshalUtil, optionalTargetObject ...*Identity) (result *Identity, err error) {
	// determine the target object that will hold the unmarshaled information
	switch len(optionalTargetObject) {
	case 0:
		result = &Identity{}
	case 1:
		result = optionalTargetObject[0]
	default:
		panic("too many arguments in call to Parse")
	}

	result.publicKey, err = ed255192.ParsePublicKey(marshalUtil)
	if err != nil {
		return
	}
	result.id = NewID(result.publicKey)

	return
}

func (i Identity) ID() ID {
	return i.id
}

func (i Identity) PublicKey() ed255192.PublicKey {
	return i.publicKey
}

func GenerateIdentity() *Identity {
	publicKey, _, err := ed255192.GenerateKey()
	if err != nil {
		panic(err)
	}

	return &Identity{
		id:        NewID(publicKey),
		publicKey: publicKey,
	}
}

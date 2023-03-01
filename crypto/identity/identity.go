// Package identity implements a node's identity, consisting out of id and public key.
// A LocalIdentity additionally contains the private key and enables signing messages.
package identity

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
)

// LocalIdentity is a node's identity.
type Identity struct {
	id        ID                // comparable node identifier
	publicKey ed25519.PublicKey // public key used to verify signatures
}

// New creates a new identity from the given PublicKey.
func New(publicKey ed25519.PublicKey) *Identity {
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

	result.publicKey, err = ed25519.ParsePublicKey(marshalUtil)
	if err != nil {
		return
	}
	result.id = NewID(result.publicKey)

	return
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

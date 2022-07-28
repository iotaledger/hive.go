package identity

import (
	ed255192 "github.com/iotaledger/hive.go/core/crypto/ed25519"
)

// LocalIdentity is a local node's identity.
type LocalIdentity struct {
	*Identity
	privateKey ed255192.PrivateKey
}

// NewLocalIdentity creates a new LocalIdentity.
func NewLocalIdentity(publicKey ed255192.PublicKey, privateKey ed255192.PrivateKey) *LocalIdentity {
	return &LocalIdentity{
		Identity:   New(publicKey),
		privateKey: privateKey,
	}
}

// NewLocalIdentityWithIdentity creates a new LocalIdentity with a given LocalIdentity.
func NewLocalIdentityWithIdentity(identity *Identity, privateKey ed255192.PrivateKey) *LocalIdentity {
	return &LocalIdentity{
		Identity:   identity,
		privateKey: privateKey,
	}
}

// Sign signs the message with the local identity's private key and returns a signature.
func (l LocalIdentity) Sign(message []byte) ed255192.Signature {
	return l.privateKey.Sign(message)
}

func GenerateLocalIdentity() *LocalIdentity {
	publicKey, privateKey, err := ed255192.GenerateKey()
	if err != nil {
		panic(err)
	}

	return &LocalIdentity{
		Identity:   New(publicKey),
		privateKey: privateKey,
	}
}

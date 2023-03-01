package identity

import (
	"crypto/sha256"
	"testing"

	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/lo"
)

func TestID(t *testing.T) {
	pub, _, err := ed25519.GenerateKey()
	require.NoError(t, err)

	id := NewID(pub)

	bytes := sha256.Sum256(lo.PanicOnErr(pub.Bytes()))

	assert.Equal(t, lo.PanicOnErr(id.Bytes()), bytes[:])
	assert.Equal(t, id.String(), base58.Encode(bytes[:])[:8])
}

func TestNewIdentity(t *testing.T) {
	pub, _, err := ed25519.GenerateKey()
	require.NoError(t, err)

	identity := New(pub)
	ref := &Identity{
		id:        NewID(pub),
		publicKey: pub,
	}

	assert.Equal(t, identity.PublicKey(), pub)
	assert.Equal(t, identity.ID(), NewID(pub))
	assert.Equal(t, identity, ref)
}

func TestNewLocalIdentity(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey()
	require.NoError(t, err)

	localIdentity := NewLocalIdentity(pub, priv)

	ref := &LocalIdentity{
		Identity:   New(pub),
		privateKey: priv,
	}

	assert.Equal(t, localIdentity.PublicKey(), pub)
	assert.Equal(t, localIdentity.ID(), NewID(pub))
	assert.Equal(t, localIdentity.Sign([]byte("toSign")), priv.Sign([]byte("toSign")))
	assert.Equal(t, localIdentity, ref)
}

func TestNewLocalIdentityWithIdentity(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey()
	require.NoError(t, err)

	identity := New(pub)
	localIdentity := NewLocalIdentityWithIdentity(identity, priv)

	assert.Same(t, localIdentity.Identity, identity)
}

package peer

import (
	"github.com/iotaledger/hive.go/autopeering/arrow"
	"github.com/iotaledger/hive.go/autopeering/peer/service"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestID(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey()
	require.NoError(t, err)

	local := newLocal(priv, testIP, newTestServiceRecord(), nil)
	id := identity.NewID(pub)
	assert.Equal(t, id, local.ID())
}

func TestPublicKey(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey()
	require.NoError(t, err)

	local := newLocal(priv, testIP, newTestServiceRecord(), nil)
	assert.EqualValues(t, pub, local.PublicKey())
}

func TestAddress(t *testing.T) {
	local := newTestLocal(t, nil)

	endpoint := local.Services().Get(service.PeeringKey)
	assert.EqualValues(t, endpoint.Port(), local.Address().Port)
	assert.EqualValues(t, endpoint.Network(), local.Address().Network())
}

func TestArRow(t *testing.T) {
	p := newTestLocal(t, nil)

	s, _ := arrow.NewArRow(600, 4, p.identity.Identity, 1000)
	p.SetArRow(s)

	got := p.GetArRow()
	assert.Equal(t, s, got, "Private salt")
}

func newTestLocal(t require.TestingT, db *DB) *Local {
	var priv ed25519.PrivateKey
	var err error
	if db == nil {
		priv, err = ed25519.GeneratePrivateKey()
		require.NoError(t, err)
	} else {
		priv, err = db.LocalPrivateKey()
		require.NoError(t, err)
	}
	services := service.New()
	services.Update(service.PeeringKey, testNetwork, testPort)
	return newLocal(priv, testIP, services, db)
}

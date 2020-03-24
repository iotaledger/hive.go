package peer

import (
	"crypto/ed25519"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/autopeering/peer/service"
	"github.com/iotaledger/hive.go/autopeering/salt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestID(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	local := newLocal(PrivateKey(priv), testIP, newTestServiceRecord(), nil)
	id := PublicKey(pub).ID()
	assert.Equal(t, id, local.ID())
}

func TestPublicKey(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	local := newLocal(PrivateKey(priv), testIP, newTestServiceRecord(), nil)
	assert.EqualValues(t, pub, local.PublicKey())
}

func TestAddress(t *testing.T) {
	local := newTestLocal(t, nil)

	endpoint := local.Services().Get(service.PeeringKey)
	assert.EqualValues(t, endpoint.Port(), local.Address().Port)
	assert.EqualValues(t, endpoint.Network(), local.Address().Network())
}

func TestPrivateSalt(t *testing.T) {
	p := newTestLocal(t, nil)

	s, _ := salt.NewSalt(time.Second * 10)
	p.SetPrivateSalt(s)

	got := p.GetPrivateSalt()
	assert.Equal(t, s, got, "Private salt")
}

func TestPublicSalt(t *testing.T) {
	p := newTestLocal(t, nil)

	s, _ := salt.NewSalt(time.Second * 10)
	p.SetPublicSalt(s)

	got := p.GetPublicSalt()

	assert.Equal(t, s, got, "Public salt")
}

func newTestLocal(t require.TestingT, db *DB) *Local {
	var priv PrivateKey
	var err error
	if db == nil {
		priv, err = generatePrivateKey()
		require.NoError(t, err)
	} else {
		priv, err = db.LocalPrivateKey()
		require.NoError(t, err)
	}
	services := service.New()
	services.Update(service.PeeringKey, testNetwork, testPort)
	return newLocal(priv, testIP, services, db)
}

package peer

import (
	"testing"
	"time"

	"github.com/iotaledger/hive.go/autopeering/peer/service"
	"github.com/iotaledger/hive.go/autopeering/salt"
	"github.com/iotaledger/hive.go/signature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestID(t *testing.T) {
	pub, priv, err := signature.GenerateKey(nil)
	require.NoError(t, err)

	local := newLocal(priv, newTestServiceRecord(), nil)
	id := CreateID(pub)
	assert.Equal(t, id, local.ID())
}

func TestPublicKey(t *testing.T) {
	pub, priv, err := signature.GenerateKey(nil)
	require.NoError(t, err)

	local := newLocal(priv, newTestServiceRecord(), nil)
	assert.EqualValues(t, pub, local.PublicKey())
}

func TestAddress(t *testing.T) {
	local := newTestLocal(t, nil)

	address := local.Services().Get(service.PeeringKey).String()
	assert.EqualValues(t, address, local.Address())
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
	var priv signature.PrivateKey
	var err error
	if db == nil {
		priv, err = signature.GeneratePrivateKey()
		require.NoError(t, err)
	} else {
		priv, err = db.LocalPrivateKey()
		require.NoError(t, err)
	}
	services := service.New()
	services.Update(service.PeeringKey, testNetwork, testAddress)
	return newLocal(priv, services, db)
}

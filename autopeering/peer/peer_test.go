package peer

import (
	"net"
	"testing"

	"github.com/iotaledger/hive.go/autopeering/peer/service"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testNetwork = "udp"
	testIP      = net.IPv4zero
	testPort    = 8000
	testMessage = "Hello World!"
)

func newTestServiceRecord() *service.Record {
	services := service.New()
	services.Update(service.PeeringKey, testNetwork, testPort)

	return services
}

func newTestPeer(name string) *Peer {
	key := ed25519.PublicKey{}
	copy(key[:], name)
	return NewPeer(identity.NewIdentity(key), testIP, newTestServiceRecord())
}

func TestNoServicePeer(t *testing.T) {
	key := ed25519.PublicKey{}
	services := service.New()

	assert.Panics(t, func() {
		_ = NewPeer(identity.NewIdentity(key), testIP, services)
	})
}

func TestInvalidServicePeer(t *testing.T) {
	key := ed25519.PublicKey{}
	services := service.New()
	services.Update(service.FPCKey, "network", 8001)

	assert.Panics(t, func() {
		_ = NewPeer(identity.NewIdentity(key), testIP, services)
	})
}

func TestMarshalUnmarshal(t *testing.T) {
	p := newTestPeer("test")
	data, err := p.Marshal()
	require.NoError(t, err)

	got, err := Unmarshal(data)
	require.NoError(t, err)

	assert.Equal(t, p, got)
}

func TestRecoverKeyFromSignedData(t *testing.T) {
	msg := []byte(testMessage)

	pub, priv, err := ed25519.GenerateKey()
	require.NoError(t, err)

	sig := priv.Sign(msg)

	d := signedData{pub: pub.Bytes(), msg: msg, sig: sig.Bytes()}
	key, err := RecoverKeyFromSignedData(d)
	require.NoError(t, err)

	assert.Equal(t, pub, key)
}

type signedData struct {
	pub, msg, sig []byte
}

func (d signedData) GetPublicKey() []byte { return d.pub }
func (d signedData) GetData() []byte      { return d.msg }
func (d signedData) GetSignature() []byte { return d.sig }

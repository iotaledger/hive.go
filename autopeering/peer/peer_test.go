package peer

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/izuc/zipp.foundation/autopeering/peer/service"
	"github.com/izuc/zipp.foundation/crypto/ed25519"
	"github.com/izuc/zipp.foundation/crypto/identity"
	"github.com/izuc/zipp.foundation/lo"
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

	return NewPeer(identity.New(key), testIP, newTestServiceRecord())
}

func TestNoServicePeer(t *testing.T) {
	key := ed25519.PublicKey{}
	services := service.New()

	assert.Panics(t, func() {
		_ = NewPeer(identity.New(key), testIP, services)
	})
}

func TestInvalidServicePeer(t *testing.T) {
	key := ed25519.PublicKey{}
	services := service.New()
	services.Update(service.FPCKey, "network", 8001)

	assert.Panics(t, func() {
		_ = NewPeer(identity.New(key), testIP, services)
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

func TestMarshalUnmarshalJSON(t *testing.T) {
	p := newTestPeer("test")

	data, err := json.Marshal(p)
	require.NoError(t, err)

	got := &Peer{}
	err = json.Unmarshal(data, got)
	require.NoError(t, err)

	assert.Equal(t, p, got)
}

func TestMarshalUnmarshalJSONSlice(t *testing.T) {
	peers := []*Peer{newTestPeer("test2"), newTestPeer("test2"), newTestPeer("test2")}

	data, err := json.Marshal(peers)
	require.NoError(t, err)

	var got []*Peer
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, peers, got)
}

func TestRecoverKeyFromSignedData(t *testing.T) {
	msg := []byte(testMessage)

	pub, priv, err := ed25519.GenerateKey()
	require.NoError(t, err)

	sig := priv.Sign(msg)

	d := signedData{pub: lo.PanicOnErr(pub.Bytes()), msg: msg, sig: lo.PanicOnErr(sig.Bytes())}
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

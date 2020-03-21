package peer

import (
	"testing"

	"github.com/iotaledger/hive.go/autopeering/peer/service"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testNetwork = "udp"
	testAddress = "127.0.0.1:8000"
	testMessage = "Hello World!"
)

func newTestServiceRecord() *service.Record {
	services := service.New()
	services.Update(service.PeeringKey, testNetwork, testAddress)

	return services
}

func newTestPeer(name string) *Peer {
	key := ed25519.PublicKey{}
	copy(key[:], name)
	return NewPeer(key, newTestServiceRecord())
}

func TestNoServicePeer(t *testing.T) {
	key := ed25519.PublicKey{}
	services := service.New()

	assert.Panics(t, func() {
		_ = NewPeer(key, services)
	})
}

func TestInvalidServicePeer(t *testing.T) {
	key := ed25519.PublicKey{}
	services := service.New()
	services.Update(service.FPCKey, "network", "address")

	assert.Panics(t, func() {
		_ = NewPeer(key, services)
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

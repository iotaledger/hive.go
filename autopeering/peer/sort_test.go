package peer

import (
	"github.com/iotaledger/hive.go/autopeering/arrow"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/identity"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func newTestPeerWithID(ID byte) *Peer {
	var key ed25519.PublicKey
	key[0] = ID
	return NewPeer(identity.New(key), net.IPv4zero, newTestServiceRecord())
}

func TestOrderedDistanceList(t *testing.T) {
	type testCase struct {
		anchor  []byte
		salt    []byte
		ordered bool
	}

	tests := []testCase{
		{
			anchor:  []byte("X"),
			salt:    []byte("salt"),
			ordered: true,
		},
	}

	remotePeers := make([]*Peer, 10)
	for i := range remotePeers {
		remotePeers[i] = newTestPeerWithID(byte(i + 1))
	}
	localArRow, _ := arrow.NewArRow(100, 4, remotePeers[0].Identity, 1000)

	for _, test := range tests {
		d := SortByInbound(0, localArRow, remotePeers, 1000)
		prev := d[0]
		for _, next := range d[1:] {
			got := prev.Distance < next.Distance
			assert.Equal(t, test.ordered, got, "Ordered distance list")
			prev = next
		}
	}
}

//nolint:gosec // we don't care about these linters in test cases
package peer

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/izuc/zipp.foundation/crypto/ed25519"
	"github.com/izuc/zipp.foundation/crypto/identity"
)

func newTestPeerWithID(id byte) *Peer {
	var key ed25519.PublicKey
	key[0] = id

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

	for _, test := range tests {
		d := SortBySalt(test.anchor, test.salt, remotePeers)
		prev := d[0]
		for _, next := range d[1:] {
			got := prev.Distance < next.Distance
			assert.Equal(t, test.ordered, got, "Ordered distance list")
			prev = next
		}
	}
}

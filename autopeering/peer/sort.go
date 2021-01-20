package peer

import (
	"github.com/iotaledger/hive.go/autopeering/arrow"
	"github.com/iotaledger/hive.go/autopeering/distance"
	"sort"
	"time"
)

// PeerDistance defines the relative distance wrt a remote peer
type PeerDistance struct {
	Remote   *Peer
	Channel  int
	Distance float64
}

// byDistance is a slice of PeerDistance used to sort
type byDistance []PeerDistance

func (a byDistance) Len() int           { return len(a) }
func (a byDistance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byDistance) Less(i, j int) bool { return a[i].Distance < a[j].Distance }

// NewPeerDistance returns a new PeerDistance
func NewPeerDistance(localAr, remoteAr float64, channel int, remote *Peer) PeerDistance {
	return PeerDistance{
		Remote:   remote,
		Channel:  channel,
		Distance: distance.ByArs(localAr, remoteAr),
	}
}

// SortByOutbound returns a slice of PeerDistance given a list of remote peers sorted from outbound perspective
func SortByOutbound(channel int, localArRow *arrow.ArRow, remotePeers []*Peer, epoch uint64) (result []PeerDistance) {
	result = make(byDistance, len(remotePeers))

	for i, remote := range remotePeers {
		peerRows, _ := arrow.NewArRow(time.Until(localArRow.GetExpiration()), len(localArRow.GetArs()), remote.Identity, epoch)
		result[i] = NewPeerDistance(localArRow.GetArs()[channel], peerRows.GetRows()[channel], channel, remote)
	}
	sort.Sort(byDistance(result))
	return result
}

// SortByInbound returns a slice of PeerDistance given a list of remote peers sorted from inbound perspective
func SortByInbound(channel int, localArs *arrow.ArRow, remotePeers []*Peer, epoch uint64) (result []PeerDistance) {
	result = make(byDistance, len(remotePeers))

	for i, remote := range remotePeers {

		peerRows, _ := arrow.NewArRow(time.Until(localArs.GetExpiration()), len(localArs.GetRows()), remote.Identity, epoch)
		result[i] = NewPeerDistance(localArs.GetRows()[channel], peerRows.GetArs()[channel], channel, remote)
	}
	sort.Sort(byDistance(result))
	return result
}

package peer

import (
	"github.com/iotaledger/hive.go/autopeering/arrow"
	"sort"
	"time"

	"github.com/iotaledger/hive.go/autopeering/distance"
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

// SortBySalt returns a slice of PeerDistance given a list of remote peers
func SortByOutbound(channel int, localArRow *arrow.ArRow, remotePeers []*Peer) (result []PeerDistance) {
	result = make(byDistance, len(remotePeers))
	for i, remote := range remotePeers {
		peerRows, _ := arrow.NewArRow(localArRow.GetExpiration().Sub(time.Now()), len(localArRow.GetArs()), remote.Identity)
		result[i] = NewPeerDistance(localArRow.GetArs()[channel], peerRows.GetRows()[channel], channel, remote)
	}
	sort.Sort(byDistance(result))
	return result
}

// SortBySalt returns a slice of PeerDistance given a list of remote peers
func SortByInbound(channel int, localArs *arrow.ArRow, remotePeers []*Peer) (result []PeerDistance) {
	result = make(byDistance, len(remotePeers))
	for i, remote := range remotePeers {
		peerRows, _ := arrow.NewArRow(localArs.GetExpiration().Sub(time.Now()), len(localArs.GetRows()), remote.Identity)
		result[i] = NewPeerDistance(localArs.GetRows()[channel], peerRows.GetArs()[channel], channel, remote)
	}
	sort.Sort(byDistance(result))
	return result
}

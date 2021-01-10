package peer

import (
	"github.com/iotaledger/hive.go/autopeering/ars"
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
func SortByArs(channel int, localArs *ars.Ars, remotePeers []*Peer) (result []PeerDistance) {
	result = make(byDistance, len(remotePeers))
	for i, remote := range remotePeers {
		peerArs, _ := ars.NewArs(localArs.GetExpiration().Sub(time.Now()), len(localArs.GetArs()), remote.Identity)
		result[i] = NewPeerDistance(localArs.GetArs()[channel], peerArs.GetArs()[channel], channel, remote)
	}
	sort.Sort(byDistance(result))
	return result
}

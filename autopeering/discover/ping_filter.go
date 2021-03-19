package discover

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/autopeering/server"
)

// pingFilter is the mapping of a peer and the time of its last ping packet.
type pingFilter struct {
	lastPing map[string]time.Time
	sync.RWMutex
}

func newPingFilter() *pingFilter {
	return &pingFilter{
		lastPing: make(map[string]time.Time),
	}
}

func (p *pingFilter) update(peer string, pingTime time.Time) {
	p.Lock()
	defer p.Unlock()

	p.lastPing[peer] = pingTime
}

func (p *pingFilter) delete(peer string) {
	p.Lock()
	defer p.Unlock()

	if _, exist := p.lastPing[peer]; !exist {
		return
	}

	delete(p.lastPing, peer)
}

func (p *pingFilter) validPing(peer string, pingTime time.Time) bool {
	p.RLock()
	defer p.RUnlock()

	peerLastPing, exist := p.lastPing[peer]
	if !exist {
		return true
	}

	return pingTime.Sub(peerLastPing) > server.PacketExpiration
}

package discover

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/autopeering/server"
)

var (
	blacklistThreshold = 5
)

// pingFilter is the mapping of a peer and the time of its last ping packet.
type pingFilter struct {
	lastPing map[string]history
	sync.RWMutex
}

type history struct {
	t       time.Time
	counter int
}

func newPingFilter() *pingFilter {
	return &pingFilter{
		lastPing: make(map[string]history),
	}
}

func (p *pingFilter) update(peer string, pingTime time.Time) {
	p.Lock()
	defer p.Unlock()

	p.lastPing[peer] = history{pingTime, 0}
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
	p.Lock()
	defer p.Unlock()

	peerLastPing, exist := p.lastPing[peer]
	if !exist {
		return true
	}

	valid := pingTime.Sub(peerLastPing.t) > server.PacketExpiration
	if !valid {
		peerLastPing.counter++
		p.lastPing[peer] = peerLastPing
	}

	return valid
}

func (p *pingFilter) blacklist(peer string) bool {
	p.RLock()
	defer p.RUnlock()

	return p.lastPing[peer].counter > blacklistThreshold
}

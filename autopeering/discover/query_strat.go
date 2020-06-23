package discover

import (
	"container/ring"
	"math/rand"
	"sync"
	"time"
)

// doQuery is the main method of the query strategy.
// It writes the next time this function should be called by the manager to next.
// The current strategy is to always select the latest verified peer and one of
// the peers that returned the most number of peers the last time it was queried.
func (m *manager) doQuery(next chan<- time.Duration) {
	defer func() { next <- queryInterval }()

	ps := m.peersToQuery()
	if len(ps) == 0 {
		return
	}
	m.log.Debugw("querying",
		"#peers", len(ps),
	)

	// request from peers in parallel
	var wg sync.WaitGroup
	wg.Add(len(ps))
	for _, p := range ps {
		go m.requestWorker(p, &wg)
	}
	wg.Wait()
}

func (m *manager) requestWorker(mp *mpeer, wg *sync.WaitGroup) {
	defer wg.Done()

	p := unwrapPeer(mp)
	peers, err := m.net.DiscoveryRequest(p)
	if err != nil {
		m.log.Debugw("query failed",
			"peer", mp,
			"err", err,
		)
		m.deletePeer(p.ID())
		return
	}

	var added uint32
	for _, rp := range peers {
		if m.addDiscoveredPeer(rp) {
			added++
		}
	}
	mp.lastNewPeers.Store(added)

	m.log.Debugw("queried",
		"peer", mp,
		"#added", added,
	)
}

// peersToQuery selects the peers that should be queried.
func (m *manager) peersToQuery() []*mpeer {
	ps := m.verifiedPeers()
	if len(ps) == 0 {
		return nil
	}

	latest := ps[0]
	if len(ps) == 1 {
		return []*mpeer{latest}
	}

	// find the 3 heaviest peers
	r := ring.New(3)
	for i, p := range ps {
		if i == 0 {
			continue // the latest peer is already included
		}
		if r.Value == nil {
			r.Value = p
		} else if p.lastNewPeers.Load() >= r.Value.(*mpeer).lastNewPeers.Load() {
			r = r.Next()
			r.Value = p
		}
	}

	// select a random peer from the heaviest ones
	r.Move(rand.Intn(r.Len()))

	return []*mpeer{latest, r.Value.(*mpeer)}
}

package server

import "sync"

type blacklist struct {
	list map[string]bool
	sync.RWMutex
}

func newBlacklist() *blacklist {
	return &blacklist{
		list: make(map[string]bool),
	}
}

func (b *blacklist) Add(peer string) {
	b.Lock()
	defer b.Unlock()

	b.list[peer] = true
}

func (b *blacklist) Load(peer string) bool {
	b.RLock()
	defer b.RUnlock()

	if _, exist := b.list[peer]; !exist {
		return false
	}

	return true
}

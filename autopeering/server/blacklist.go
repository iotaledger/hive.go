package server

import (
	"sync"
)

type blacklist struct {
	list map[string]bool
	sync.RWMutex
}

func newBlacklist() *blacklist {
	return &blacklist{
		list: make(map[string]bool),
	}
}

func (b *blacklist) Add(peer string) bool {
	b.Lock()
	defer b.Unlock()

	if peer != "" {
		b.list[peer] = true
		return true
	}

	return false
}

func (b *blacklist) PeerExist(peer string) bool {
	b.RLock()
	defer b.RUnlock()

	if peer != "" {
		_, existence := b.list[peer]
		return existence
	}
	return false

	//if _, exist := b.list[peer]; !exist {
	//	return false
	//}
	//
	//return true
}

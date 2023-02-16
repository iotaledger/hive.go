package mapdb

import (
	"strings"
	"sync"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/utils"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
)

type syncedKVMap struct {
	sync.RWMutex
	m map[string][]byte
}

func (s *syncedKVMap) has(key []byte) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.m[string(key)]

	return ok
}

func (s *syncedKVMap) get(key []byte) ([]byte, bool) {
	s.RLock()
	defer s.RUnlock()
	value, ok := s.m[string(key)]
	if !ok {
		return nil, false
	}
	// always copy the value
	return byteutils.ConcatBytes(value), true
}

func (s *syncedKVMap) set(key, value []byte) {
	s.Lock()
	defer s.Unlock()
	// always copy the value
	s.m[string(key)] = byteutils.ConcatBytes(value)
}

func (s *syncedKVMap) delete(key []byte) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, string(key))
}

func (s *syncedKVMap) deletePrefix(keyPrefix []byte) {
	s.Lock()
	defer s.Unlock()
	prefix := string(keyPrefix)
	for key := range s.m {
		if strings.HasPrefix(key, prefix) {
			delete(s.m, key)
		}
	}
}

func (s *syncedKVMap) iterate(realm []byte, keyPrefix []byte, consume func(key, value []byte) bool, iterDirection ...kvstore.IterDirection) {
	// take a snapshot of the current elements
	s.RLock()
	copiedElements := make(map[string][]byte)
	prefix := byteutils.ConcatBytesToString(realm, keyPrefix)
	for key, value := range s.m {
		if strings.HasPrefix(key, prefix) {
			copiedElements[key] = byteutils.ConcatBytes(value)
		}
	}
	s.RUnlock()

	keysSlice := make([]string, 0, len(copiedElements))
	for k := range copiedElements {
		keysSlice = append(keysSlice, k)
	}

	// iterate through found elements
	for _, key := range utils.SortSlice(keysSlice, iterDirection...) {
		if !consume([]byte(key)[len(realm):], copiedElements[key]) {
			break
		}
	}
}

func (s *syncedKVMap) iterateKeys(realm []byte, keyPrefix []byte, consume func(key []byte) bool, iterDirection ...kvstore.IterDirection) {
	// take a snapshot of the current elements
	s.RLock()
	copiedElements := make(map[string]struct{})
	prefix := byteutils.ConcatBytesToString(realm, keyPrefix)
	for key := range s.m {
		if strings.HasPrefix(key, prefix) {
			copiedElements[key] = struct{}{}
		}
	}
	s.RUnlock()

	keysSlice := make([]string, 0, len(copiedElements))
	for k := range copiedElements {
		keysSlice = append(keysSlice, k)
	}

	// iterate through found elements
	for _, key := range utils.SortSlice(keysSlice, iterDirection...) {
		if !consume([]byte(key)[len(realm):]) {
			break
		}
	}
}

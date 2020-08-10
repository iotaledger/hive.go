package mapdb

import (
	"strings"
	"sync"
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
	return append([]byte{}, value...), true
}

func (s *syncedKVMap) set(key, value []byte) {
	s.Lock()
	defer s.Unlock()
	// always copy the value
	s.m[string(key)] = append([]byte{}, value...)
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

func (s *syncedKVMap) iterate(realm []byte, keyPrefix []byte, consume func(key, value []byte) bool) {
	s.RLock()
	defer s.RUnlock()
	prefix := string(append(realm, keyPrefix...))
	for key, value := range s.m {
		if strings.HasPrefix(key, prefix) {
			if !consume([]byte(key)[len(realm):], append([]byte{}, value...)) {
				break
			}
		}
	}
}

func (s *syncedKVMap) iterateKeys(realm []byte, keyPrefix []byte, consume func(key []byte) bool) {
	s.RLock()
	defer s.RUnlock()
	prefix := string(append(realm, keyPrefix...))
	for key := range s.m {
		if strings.HasPrefix(key, prefix) {
			if !consume([]byte(key)[len(realm):]) {
				break
			}
		}
	}
}

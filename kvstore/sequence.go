package kvstore

import (
	"encoding/binary"
	"sync"

	"github.com/iotaledger/hive.go/ierrors"
)

// Sequence represents a simple integer sequence backed by a KVStore.
// A Sequence can be used to get a list of monotonically increasing integers.
type Sequence struct {
	sync.Mutex
	store    KVStore
	key      []byte
	interval uint64 // the interval in which numbers are backed by the store
	next     uint64 // next number to be returned
	reserved uint64 // until which number is reserved in the store
}

// NewSequence initiates a new sequence object backed by the provided store.
// The interval value defines how many Next() requests can be served from memory without an access to the store.
// Multiple sequences can be created by providing different keys.
func NewSequence(store KVStore, key []byte, interval uint64) (*Sequence, error) {
	if interval == 0 {
		panic("interval must be greater than zero")
	}

	seq := &Sequence{
		store:    store,
		key:      key,
		next:     0,
		reserved: 0,
		interval: interval,
	}

	return seq, nil
}

// Next returns the next integer in the sequence.
func (seq *Sequence) Next() (uint64, error) {
	seq.Lock()
	defer seq.Unlock()

	if seq.next >= seq.reserved {
		if err := seq.update(); err != nil {
			return 0, err
		}
	}

	val := seq.next
	seq.next++

	return val, nil
}

// Release the leased sequence to avoid wasted integers.
func (seq *Sequence) Release() error {
	seq.Lock()
	defer seq.Unlock()

	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], seq.next)
	if err := seq.store.Set(seq.key, buf[:]); err != nil {
		return err
	}

	seq.reserved = seq.next

	return nil
}

func (seq *Sequence) update() error {
	value, err := seq.store.Get(seq.key)
	switch {
	case ierrors.Is(err, ErrKeyNotFound):
		seq.next = 0
	case err != nil:
		return err
	default:
		num := binary.BigEndian.Uint64(value)
		seq.next = num
	}

	// reserve the interval and set in store
	reserved := seq.next + seq.interval
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], reserved)
	err = seq.store.Set(seq.key, buf[:])
	if err != nil {
		return err
	}
	seq.reserved = reserved

	return nil
}

package boltdb

import (
	"bytes"
	"sync"

	"github.com/iotaledger/hive.go/objectstorage"
	"go.etcd.io/bbolt"
)

// New creates a new Storage with the underlying BoltDB.
func New(db *bbolt.DB) *Storage {
	return &Storage{instance: db}
}

// Storage implements the ObjectStorage Storage interface around a BoltDB instance.
type Storage struct {
	bucket   []byte
	instance *bbolt.DB
}

func buildPrefixedKey(prefixes [][]byte) []byte {
	var prefix []byte
	for _, p := range prefixes {
		prefix = append(prefix, p...)
	}
	return prefix
}

func copyBytes(source []byte) []byte {
	cpy := make([]byte, len(source))
	copy(cpy, source)
	return cpy
}

func (s Storage) iterate(realm []byte, prefixes [][]byte, copyValues bool, kvConsumerFunc objectstorage.IteratorKeyValueConsumerFunc) error {
	return s.instance.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(realm)
		if b == nil {
			return nil
		}
		c := b.Cursor()

		if len(prefixes) > 0 {
			prefix := buildPrefixedKey(prefixes)
			for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
				val := v
				if copyValues {
					val = copyBytes(v)
				}
				if !kvConsumerFunc(copyBytes(k), val) {
					break
				}
			}
			return nil
		}

		for k, v := c.First(); k != nil; k, v = c.Next() {
			val := v
			if copyValues {
				val = copyBytes(v)
			}
			if !kvConsumerFunc(copyBytes(k), val) {
				break
			}
		}
		return nil
	})
}

func (s *Storage) Iterate(realm []byte, prefixes [][]byte, _ bool, kvConsumerFunc objectstorage.IteratorKeyValueConsumerFunc) error {
	return s.iterate(realm, prefixes, true, kvConsumerFunc)
}

func (s *Storage) IterateKeys(realm []byte, prefixes [][]byte, consumerFunc objectstorage.IteratorKeyConsumerFunc) error {
	// same as with values but we simply don't copy them
	return s.iterate(realm, prefixes, false, func(key []byte, _ []byte) bool {
		return consumerFunc(key)
	})
}

func (s *Storage) Clear(realm []byte) error {
	return s.instance.Update(func(tx *bbolt.Tx) error {
		if tx.Bucket(realm) == nil {
			return nil
		}
		return tx.DeleteBucket(realm)
	})
}

func (s *Storage) Get(realm []byte, key []byte) ([]byte, error) {
	var val []byte
	if err := s.instance.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(realm)
		if b == nil {
			return nil
		}
		val = b.Get(key)
		return nil
	}); err != nil {
		return nil, err
	}
	if val == nil {
		return nil, objectstorage.ErrKeyNotFound
	}
	return val, nil
}

func (s *Storage) Set(realm []byte, key []byte, value []byte) error {
	return s.instance.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(realm)
		if err != nil {
			return err
		}
		return b.Put(key, value)
	})
}

func (s *Storage) Has(realm []byte, key []byte) (bool, error) {
	var has bool
	err := s.instance.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(realm)
		if b == nil {
			return nil
		}
		has = b.Get(key) != nil
		return nil
	})
	return has, err
}

func (s *Storage) Delete(realm []byte, key []byte) error {
	return s.instance.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(realm)
		if b == nil {
			return objectstorage.ErrKeyNotFound
		}
		if err := b.Delete(key); err != nil {
			return objectstorage.ErrKeyNotFound
		}
		return nil
	})
}

func (s *Storage) Batched() objectstorage.BatchedMutations {
	// we don't use BoltDB's Batch(), because it basically is only
	// a way to let BoltDB decide how to make a batched update itself,
	// which is only useful if Batch() is called from multiple goroutines.
	// instead, if we collect the mutations and then do a single
	// update, we have the batched mutations we actually want.
	return &BatchedMutations{instance: s.instance}
}

type realmkvtuple struct {
	realm []byte
	key   []byte
	value []byte
}

// BatchedMutations is a wrapper to do a batched update on a BoltDB.
type BatchedMutations struct {
	sync.Mutex
	instance *bbolt.DB
	sets     []realmkvtuple
	deletes  []realmkvtuple
}

func (b *BatchedMutations) Set(realm []byte, key []byte, value []byte) error {
	b.Lock()
	defer b.Unlock()
	b.sets = append(b.sets, realmkvtuple{realm, key, value})
	return nil
}

func (b *BatchedMutations) Delete(realm []byte, key []byte) error {
	b.Lock()
	defer b.Unlock()
	b.deletes = append(b.deletes, realmkvtuple{realm, key, nil})
	return nil
}

func (b *BatchedMutations) Cancel() {
	// do nothing
}

func (b *BatchedMutations) Commit() error {
	return b.instance.Update(func(tx *bbolt.Tx) error {
		for i := 0; i < len(b.sets); i++ {
			bucket, err := tx.CreateBucketIfNotExists(b.sets[i].realm)
			if err != nil {
				return err
			}
			if err := bucket.Put(b.sets[i].key, b.sets[i].value); err != nil {
				return err
			}
		}
		for i := 0; i < len(b.deletes); i++ {
			bucket := tx.Bucket(b.sets[i].realm)
			if bucket == nil {
				continue
			}
			if err := bucket.Delete(b.sets[i].key); err != nil {
				return err
			}
		}
		return nil
	})
}

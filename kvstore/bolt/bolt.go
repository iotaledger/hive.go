package bolt

import (
	"bytes"
	"math"
	"sync"

	"go.etcd.io/bbolt"

	"github.com/iotaledger/hive.go/kvstore"
)

const (
	MaxBoltBatchSize = 50_000
)

// KVStore implements the KVStore interface around a BoltDB instance.
type boltStore struct {
	instance *bbolt.DB
	bucket   []byte
}

// New creates a new KVStore with the underlying BoltDB.
func New(db *bbolt.DB) kvstore.KVStore {
	return &boltStore{
		instance: db,
	}
}

func (s *boltStore) WithRealm(realm kvstore.Realm) kvstore.KVStore {
	return &boltStore{
		instance: s.instance,
		bucket:   realm,
	}
}

func (s *boltStore) Realm() kvstore.Realm {
	return s.bucket
}

func buildPrefixedKey(prefixes []kvstore.KeyPrefix) []byte {
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

func (s boltStore) iterate(prefixes []kvstore.KeyPrefix, copyValues bool, kvConsumerFunc kvstore.IteratorKeyValueConsumerFunc) error {
	return s.instance.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.bucket)
		if b == nil {
			return nil
		}
		c := b.Cursor()

		if len(prefixes) > 0 {
			prefix := buildPrefixedKey(prefixes)
			for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
				value := v
				if copyValues {
					value = copyBytes(v)
				}
				if !kvConsumerFunc(copyBytes(k), value) {
					break
				}
			}
			return nil
		}

		for k, v := c.First(); k != nil; k, v = c.Next() {
			value := v
			if copyValues {
				value = copyBytes(v)
			}
			if !kvConsumerFunc(copyBytes(k), value) {
				break
			}
		}
		return nil
	})
}

func (s *boltStore) Iterate(prefixes []kvstore.KeyPrefix, kvConsumerFunc kvstore.IteratorKeyValueConsumerFunc) error {
	return s.iterate(prefixes, true, kvConsumerFunc)
}

func (s *boltStore) IterateKeys(prefixes []kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc) error {
	// same as with values but we simply don't copy them
	return s.iterate(prefixes, false, func(key kvstore.Key, _ kvstore.Value) bool {
		return consumerFunc(key)
	})
}

func (s *boltStore) Clear() error {
	return s.instance.Update(func(tx *bbolt.Tx) error {
		if tx.Bucket(s.bucket) == nil {
			return nil
		}
		return tx.DeleteBucket(s.bucket)
	})
}

func (s *boltStore) Get(key kvstore.Key) (kvstore.Value, error) {
	var value []byte
	if err := s.instance.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.bucket)
		if b == nil {
			return nil
		}
		v := b.Get(key)
		if v != nil {
			value = copyBytes(v)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if value == nil {
		return nil, kvstore.ErrKeyNotFound
	}
	return value, nil
}

func (s *boltStore) Set(key kvstore.Key, value kvstore.Value) error {
	return s.instance.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(s.bucket)
		if err != nil {
			return err
		}
		return b.Put(key, value)
	})
}

func (s *boltStore) Has(key kvstore.Key) (bool, error) {
	var has bool
	err := s.instance.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.bucket)
		if b == nil {
			return nil
		}
		has = b.Get(key) != nil
		return nil
	})
	return has, err
}

func (s *boltStore) Delete(key kvstore.Key) error {
	return s.instance.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(s.bucket)
		if err != nil {
			return err
		}
		if err := b.Delete(key); err != nil {
			return kvstore.ErrKeyNotFound
		}
		return nil
	})
}

func (s *boltStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	return s.instance.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(s.bucket)
		if err != nil {
			return err
		}
		c := b.Cursor()
		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if err := b.Delete(k); err != nil {
				return kvstore.ErrKeyNotFound
			}
		}
		return nil
	})
}

func (s *boltStore) Batched() kvstore.BatchedMutations {
	// we don't use BoltDB's Batch(), because it basically is only
	// a way to let BoltDB decide how to make a batched update itself,
	// which is only useful if Batch() is called from multiple goroutines.
	// instead, if we collect the mutations and then do a single
	// update, we have the batched mutations we actually want.
	return &batchedMutations{
		instance: s.instance,
		bucket:   s.bucket,
	}
}

type kvtuple struct {
	key   kvstore.Key
	value kvstore.Value
}

// batchedMutations is a wrapper to do a batched update on a BoltDB.
type batchedMutations struct {
	sync.Mutex
	instance *bbolt.DB
	bucket   []byte
	sets     []kvtuple
	deletes  []kvtuple
}

func (b *batchedMutations) Set(key kvstore.Key, value kvstore.Value) error {
	b.Lock()
	defer b.Unlock()
	b.sets = append(b.sets, kvtuple{key, value})
	return nil
}

func (b *batchedMutations) Delete(key kvstore.Key) error {
	b.Lock()
	defer b.Unlock()
	b.deletes = append(b.deletes, kvtuple{key, nil})
	return nil
}

func (b *batchedMutations) Cancel() {
	// do nothing
}

func (b *batchedMutations) subBatchOperation(batchOpcount int32, operation func(opIndex int32, bucket *bbolt.Bucket) error) error {
	batchAmount := int(math.Ceil(float64(batchOpcount) / float64(MaxBoltBatchSize)))
	for i := 0; i < batchAmount; i++ {

		batchStart := int32(i * MaxBoltBatchSize)
		batchEnd := batchStart + int32(MaxBoltBatchSize)

		if batchEnd > batchOpcount {
			batchEnd = batchOpcount
		}

		err := b.instance.Update(func(tx *bbolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists(b.bucket)
			if err != nil {
				return err
			}
			for j := batchStart; j < batchEnd; j++ {
				if err := operation(j, bucket); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *batchedMutations) Commit() error {
	b.Lock()
	defer b.Unlock()

	setCount := int32(len(b.sets))
	deleteCount := int32(len(b.deletes))

	err := b.subBatchOperation(setCount, func(opIndex int32, bucket *bbolt.Bucket) error {
		if err := bucket.Put(b.sets[opIndex].key, b.sets[opIndex].value); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = b.subBatchOperation(deleteCount, func(opIndex int32, bucket *bbolt.Bucket) error {
		if err := bucket.Delete(b.deletes[opIndex].key); err != nil {
			return err
		}
		return nil
	})
	return err
}

package bolt

import (
	"bytes"
	"errors"
	"sync"

	"go.etcd.io/bbolt"

	"github.com/iotaledger/hive.go/v2/byteutils"
	"github.com/iotaledger/hive.go/v2/kvstore"
	"github.com/iotaledger/hive.go/v2/kvstore/utils"
	"github.com/iotaledger/hive.go/v2/types"
)

const (
	MaxBoltBatchSize = 50_000
)

// boltStore implements the KVStore interface around a BoltDB instance.
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
	if len(s.bucket) == 0 {
		return []byte("bolt")
	}
	return s.bucket
}

// Shutdown marks the store as shutdown.
func (s *boltStore) Shutdown() {
}

// getIterFuncs returns the function pointers for the iteration based on the given settings.
func (s *boltStore) getIterFuncs(c *bbolt.Cursor, keyPrefix []byte, iterDirection ...kvstore.IterDirection) (start func() (k []byte, v []byte), valid func(k []byte) bool, move func() (k []byte, v []byte), err error) {

	startFunc := c.First
	validFunc := func(k []byte) bool {
		return k != nil
	}
	moveFunc := c.Next

	if len(keyPrefix) > 0 {
		startFunc = func() (k []byte, v []byte) {
			return c.Seek(keyPrefix)
		}
		validFunc = func(k []byte) bool {
			return k != nil && bytes.HasPrefix(k, keyPrefix)
		}
	}

	if kvstore.GetIterDirection(iterDirection...) == kvstore.IterDirectionBackward {
		startFunc = c.Last
		moveFunc = c.Prev

		if len(keyPrefix) > 0 {
			// we need to search the first item after the prefix
			prefixUpperBound := utils.KeyPrefixUpperBound(keyPrefix)
			if prefixUpperBound == nil {
				return nil, nil, nil, errors.New("no upper bound for prefix")
			}

			startFunc = func() (k []byte, v []byte) {
				k, v = c.Seek(prefixUpperBound)

				// if the upper bound exists (not part of the prefix set), we need to use the next entry
				if !validFunc(k) {
					k, v = moveFunc()
				}
				return k, v
			}
		}
	}

	return startFunc, validFunc, moveFunc, nil
}

func (s boltStore) iterate(prefix kvstore.KeyPrefix, copyValues bool, kvConsumerFunc kvstore.IteratorKeyValueConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	return s.instance.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.Realm())
		if b == nil {
			return nil
		}
		c := b.Cursor()

		startFunc, validFunc, moveFunc, err := s.getIterFuncs(c, prefix, iterDirection...)
		if err != nil {
			return err
		}

		for k, v := startFunc(); validFunc(k); k, v = moveFunc() {
			value := v
			if copyValues {
				value = utils.CopyBytes(v)
			}
			if !kvConsumerFunc(utils.CopyBytes(k), value) {
				break
			}
		}
		return nil
	})
}

// Iterate iterates over all keys and values with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys and values.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (s *boltStore) Iterate(prefix kvstore.KeyPrefix, kvConsumerFunc kvstore.IteratorKeyValueConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	return s.iterate(prefix, true, kvConsumerFunc, iterDirection...)
}

// IterateKeys iterates over all keys with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (s *boltStore) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	// same as with values but we simply don't copy them
	return s.iterate(prefix, false, func(key kvstore.Key, _ kvstore.Value) bool {
		return consumerFunc(key)
	}, iterDirection...)
}

func (s *boltStore) Clear() error {
	return s.instance.Update(func(tx *bbolt.Tx) error {
		if tx.Bucket(s.Realm()) == nil {
			return nil
		}
		return tx.DeleteBucket(s.Realm())
	})
}

func (s *boltStore) Get(key kvstore.Key) (kvstore.Value, error) {
	var value []byte
	if err := s.instance.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.Realm())
		if b == nil {
			return nil
		}
		v := b.Get(key)
		if v != nil {
			value = utils.CopyBytes(v)
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
		b, err := tx.CreateBucketIfNotExists(s.Realm())
		if err != nil {
			return err
		}
		return b.Put(key, value)
	})
}

func (s *boltStore) Has(key kvstore.Key) (bool, error) {
	var has bool
	err := s.instance.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.Realm())
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
		b := tx.Bucket(s.Realm())
		if b == nil {
			return nil
		}
		if err := b.Delete(key); err != nil {
			return kvstore.ErrKeyNotFound
		}
		return nil
	})
}

func (s *boltStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	return s.instance.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.Realm())
		if b == nil {
			return nil
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
		kvStore:          s,
		instance:         s.instance,
		bucket:           s.Realm(),
		setOperations:    make(map[string]kvstore.Value),
		deleteOperations: make(map[string]types.Empty),
	}
}

func (s *boltStore) Flush() error {
	return s.instance.Sync()
}

func (s *boltStore) Close() error {
	return s.instance.Close()
}

// batchedMutations is a wrapper to do a batched update on a BoltDB.
type batchedMutations struct {
	kvStore *boltStore
	sync.Mutex
	instance         *bbolt.DB
	bucket           []byte
	setOperations    map[string]kvstore.Value
	deleteOperations map[string]types.Empty
}

func (b *batchedMutations) Set(key kvstore.Key, value kvstore.Value) error {
	stringKey := byteutils.ConcatBytesToString(key)

	b.Lock()
	defer b.Unlock()

	delete(b.deleteOperations, stringKey)
	b.setOperations[stringKey] = value

	return nil
}

func (b *batchedMutations) Delete(key kvstore.Key) error {
	stringKey := byteutils.ConcatBytesToString(key)

	b.Lock()
	defer b.Unlock()

	delete(b.setOperations, stringKey)
	b.deleteOperations[stringKey] = types.Void

	return nil
}

func (b *batchedMutations) Cancel() {
	b.Lock()
	defer b.Unlock()

	b.setOperations = make(map[string]kvstore.Value)
	b.deleteOperations = make(map[string]types.Empty)
}

func (b *batchedMutations) Commit() (err error) {
	b.Lock()
	defer b.Unlock()

	// while we still have operations to execute ...
	for len(b.deleteOperations) >= 1 || len(b.setOperations) >= 1 {
		// ... start transaction ...
		err = b.instance.Update(func(tx *bbolt.Tx) error {
			// ... create the bucket if it does not exist ...
			bucket, err := tx.CreateBucketIfNotExists(b.bucket)
			if err != nil {
				return err
			}

			// ... collect the operations to execute within the current batch ...
			collectedOperationsCounter := 0
			collectedSetOperations := make(map[string]kvstore.Value)
			collectedDeleteOperations := make(map[string]types.Empty)
			for key, value := range b.setOperations {
				collectedSetOperations[key] = value

				collectedOperationsCounter++

				if collectedOperationsCounter >= MaxBoltBatchSize {
					break
				}
			}
			if collectedOperationsCounter < MaxBoltBatchSize {
				for key := range b.deleteOperations {
					collectedDeleteOperations[key] = types.Void

					collectedOperationsCounter++

					if collectedOperationsCounter >= MaxBoltBatchSize {
						break
					}
				}
			}

			// ... execute the collected operations
			for key, value := range collectedSetOperations {
				if err := bucket.Put([]byte(key), value); err != nil {
					return err
				}

				delete(b.setOperations, key)
			}
			for key := range collectedDeleteOperations {
				if err := bucket.Delete([]byte(key)); err != nil {
					return err
				}

				delete(b.deleteOperations, key)
			}

			return nil
		})

		// abort if we faced an error
		if err != nil {
			return
		}
	}

	return
}

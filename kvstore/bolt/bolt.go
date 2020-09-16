package bolt

import (
	"bytes"
	"sync"

	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/types"
	"go.etcd.io/bbolt"

	"github.com/iotaledger/hive.go/kvstore"
)

const (
	MaxBoltBatchSize = 50_000
)

// boltStore implements the KVStore interface around a BoltDB instance.
type boltStore struct {
	instance                     *bbolt.DB
	bucket                       []byte
	accessCallback               kvstore.AccessCallback
	accessCallbackCommandsFilter kvstore.Command
}

// New creates a new KVStore with the underlying BoltDB.
func New(db *bbolt.DB) kvstore.KVStore {
	return &boltStore{
		instance: db,
	}
}

// AccessCallback configures the store to pass all requests to the KVStore to the given callback.
// This can for example be used for debugging and to examine what the KVStore is doing.
func (s *boltStore) AccessCallback(callback kvstore.AccessCallback, commandsFilter ...kvstore.Command) {
	var accessCallbackCommandsFilter kvstore.Command
	if len(commandsFilter) == 0 {
		accessCallbackCommandsFilter = kvstore.AllCommands
	} else {
		for _, filterCommand := range commandsFilter {
			accessCallbackCommandsFilter |= filterCommand
		}
	}

	s.accessCallback = callback
	s.accessCallbackCommandsFilter = accessCallbackCommandsFilter
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
	if s.accessCallback != nil {
		s.accessCallback(kvstore.ShutdownCommand)
	}
}

func (s boltStore) iterate(prefix kvstore.KeyPrefix, copyValues bool, kvConsumerFunc kvstore.IteratorKeyValueConsumerFunc) error {
	return s.instance.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.Realm())
		if b == nil {
			return nil
		}
		c := b.Cursor()

		if len(prefix) > 0 {
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

func (s *boltStore) Iterate(prefix kvstore.KeyPrefix, kvConsumerFunc kvstore.IteratorKeyValueConsumerFunc) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.IterateCommand) {
		s.accessCallback(kvstore.IterateCommand, prefix)
	}

	return s.iterate(prefix, true, kvConsumerFunc)
}

func (s *boltStore) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.IterateKeysCommand) {
		s.accessCallback(kvstore.IterateKeysCommand, prefix)
	}

	// same as with values but we simply don't copy them
	return s.iterate(prefix, false, func(key kvstore.Key, _ kvstore.Value) bool {
		return consumerFunc(key)
	})
}

func (s *boltStore) Clear() error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.ClearCommand) {
		s.accessCallback(kvstore.ClearCommand)
	}

	return s.instance.Update(func(tx *bbolt.Tx) error {
		if tx.Bucket(s.Realm()) == nil {
			return nil
		}
		return tx.DeleteBucket(s.Realm())
	})
}

func (s *boltStore) Get(key kvstore.Key) (kvstore.Value, error) {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.GetCommand) {
		s.accessCallback(kvstore.GetCommand, key)
	}

	var value []byte
	if err := s.instance.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.Realm())
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
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.SetCommand) {
		s.accessCallback(kvstore.SetCommand, key, value)
	}

	return s.instance.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(s.Realm())
		if err != nil {
			return err
		}
		return b.Put(key, value)
	})
}

func (s *boltStore) Has(key kvstore.Key) (bool, error) {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.HasCommand) {
		s.accessCallback(kvstore.HasCommand, key)
	}

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
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.DeleteCommand) {
		s.accessCallback(kvstore.DeleteCommand, key)
	}

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
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.DeletePrefixCommand) {
		s.accessCallback(kvstore.DeletePrefixCommand, prefix)
	}

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
	if b.kvStore.accessCallback != nil && b.kvStore.accessCallbackCommandsFilter.HasBits(kvstore.SetCommand) {
		b.kvStore.accessCallback(kvstore.SetCommand, key, value)
	}

	stringKey := byteutils.ConcatBytesToString(key)

	b.Lock()
	defer b.Unlock()

	delete(b.deleteOperations, stringKey)
	b.setOperations[stringKey] = value

	return nil
}

func (b *batchedMutations) Delete(key kvstore.Key) error {
	if b.kvStore.accessCallback != nil && b.kvStore.accessCallbackCommandsFilter.HasBits(kvstore.DeleteCommand) {
		b.kvStore.accessCallback(kvstore.DeleteCommand, key)
	}

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

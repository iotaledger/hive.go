package badger

import (
	"sync"
	"sync/atomic"

	"github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"

	"github.com/izuc/zipp.foundation/ds/types"
	"github.com/izuc/zipp.foundation/kvstore"
	"github.com/izuc/zipp.foundation/kvstore/utils"
	"github.com/izuc/zipp.foundation/serializer/v2/byteutils"
)

// badgerStore implements the KVStore interface around a BadgerDB instance.
type badgerStore struct {
	instance *badger.DB
	closed   *atomic.Bool
	dbPrefix []byte
}

// New creates a new KVStore with the underlying BadgerDB.
func New(db *badger.DB) kvstore.KVStore {
	return &badgerStore{
		instance: db,
		closed:   new(atomic.Bool),
	}
}

func (s *badgerStore) WithRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	if s.closed.Load() {
		return nil, kvstore.ErrStoreClosed
	}

	return &badgerStore{
		instance: s.instance,
		closed:   s.closed,
		dbPrefix: realm,
	}, nil
}

func (s *badgerStore) WithExtendedRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	return s.WithRealm(byteutils.ConcatBytes(s.Realm(), realm))
}

func (s *badgerStore) Realm() []byte {
	return s.dbPrefix
}

// builds a key usable for the badger instance using the realm and the given prefix.
func (s *badgerStore) buildKeyPrefix(prefix kvstore.KeyPrefix) kvstore.KeyPrefix {
	return byteutils.ConcatBytes(s.dbPrefix, prefix)
}

// getIterFuncs returns the function pointers for the iteration based on the given settings.
func (s *badgerStore) getIterFuncs(it *badger.Iterator, keyPrefix []byte, iterDirection ...kvstore.IterDirection) (start func(), valid func() bool, move func(), err error) {
	startFunc := it.Rewind
	validFunc := it.Valid
	moveFunc := it.Next

	if len(keyPrefix) > 0 {
		startFunc = func() {
			it.Seek(keyPrefix)
		}
		validFunc = func() bool {
			return it.ValidForPrefix(keyPrefix)
		}
	}

	if kvstore.GetIterDirection(iterDirection...) == kvstore.IterDirectionBackward {
		if len(keyPrefix) > 0 {
			// we need to search the first item after the prefix
			prefixUpperBound := utils.KeyPrefixUpperBound(keyPrefix)
			if prefixUpperBound == nil {
				return nil, nil, nil, errors.New("no upper bound for prefix")
			}
			startFunc = func() {
				it.Seek(prefixUpperBound)

				// if the upper bound exists (not part of the prefix set), we need to use the next entry
				if !validFunc() {
					moveFunc()
				}
			}
		}
	}

	return startFunc, validFunc, moveFunc, nil
}

// Iterate iterates over all keys and values with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys and values.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (s *badgerStore) Iterate(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyValueConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	return s.instance.View(func(txn *badger.Txn) (err error) {
		keyPrefix := s.buildKeyPrefix(prefix)

		iteratorOptions := badger.DefaultIteratorOptions
		iteratorOptions.Prefix = keyPrefix
		iteratorOptions.PrefetchValues = true
		iteratorOptions.Reverse = kvstore.GetIterDirection(iterDirection...) == kvstore.IterDirectionBackward

		it := txn.NewIterator(iteratorOptions)
		defer it.Close()

		startFunc, validFunc, moveFunc, err := s.getIterFuncs(it, keyPrefix, iterDirection...)
		if err != nil {
			return err
		}

		for startFunc(); validFunc(); moveFunc() {
			item := it.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				panic(err)
			}
			if !consumerFunc(item.KeyCopy(nil)[len(s.dbPrefix):], value) {
				break
			}
		}

		return nil
	})
}

// IterateKeys iterates over all keys with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (s *badgerStore) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	return s.instance.View(func(txn *badger.Txn) (err error) {
		keyPrefix := s.buildKeyPrefix(prefix)

		iteratorOptions := badger.DefaultIteratorOptions
		iteratorOptions.Prefix = keyPrefix
		iteratorOptions.PrefetchValues = false
		iteratorOptions.Reverse = kvstore.GetIterDirection(iterDirection...) == kvstore.IterDirectionBackward

		it := txn.NewIterator(iteratorOptions)
		defer it.Close()

		startFunc, validFunc, moveFunc, err := s.getIterFuncs(it, keyPrefix, iterDirection...)
		if err != nil {
			return err
		}

		for startFunc(); validFunc(); moveFunc() {
			if !consumerFunc(it.Item().KeyCopy(nil)[len(s.dbPrefix):]) {
				break
			}
		}

		return nil
	})
}

func (s *badgerStore) Clear() error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	return s.DeletePrefix(kvstore.EmptyPrefix)
}

func (s *badgerStore) Get(key kvstore.Key) (kvstore.Value, error) {
	if s.closed.Load() {
		return nil, kvstore.ErrStoreClosed
	}

	var value []byte
	err := s.instance.View(func(txn *badger.Txn) error {
		item, err := txn.Get(byteutils.ConcatBytes(s.dbPrefix, key))
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)

		return err
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, kvstore.ErrKeyNotFound
	}

	return value, nil
}

func (s *badgerStore) Set(key kvstore.Key, value kvstore.Value) error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	return s.instance.Update(func(txn *badger.Txn) error {
		return txn.Set(byteutils.ConcatBytes(s.dbPrefix, key), value)
	})
}

func (s *badgerStore) Has(key kvstore.Key) (bool, error) {
	if s.closed.Load() {
		return false, kvstore.ErrStoreClosed
	}

	err := s.instance.View(func(txn *badger.Txn) error {
		_, err := txn.Get(byteutils.ConcatBytes(s.dbPrefix, key))

		return err
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *badgerStore) Delete(key kvstore.Key) error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	err := s.instance.Update(func(txn *badger.Txn) error {
		return txn.Delete(byteutils.ConcatBytes(s.dbPrefix, key))
	})
	if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
		return kvstore.ErrKeyNotFound
	}

	return err
}

func (s *badgerStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	return s.instance.Update(func(txn *badger.Txn) (err error) {
		iteratorOptions := badger.DefaultIteratorOptions
		iteratorOptions.Prefix = s.buildKeyPrefix(prefix)
		iteratorOptions.PrefetchValues = false

		it := txn.NewIterator(iteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			key := it.Item().KeyCopy(nil)
			if err := txn.Delete(key); err != nil {
				panic(err)
			}
		}

		return nil
	})
}

func (s *badgerStore) Flush() error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	return s.instance.Sync()
}

func (s *badgerStore) Close() error {
	if s.closed.Swap(true) {
		// was already closed
		return nil
	}

	return s.instance.Close()
}

func (s *badgerStore) Batched() (kvstore.BatchedMutations, error) {
	if s.closed.Load() {
		return nil, kvstore.ErrStoreClosed
	}

	return &batchedMutations{
		kvStore:          s,
		store:            s.instance,
		dbPrefix:         s.dbPrefix,
		setOperations:    make(map[string]kvstore.Value),
		deleteOperations: make(map[string]types.Empty),
		closed:           s.closed,
	}, nil
}

// batchedMutations is a wrapper around a WriteBatch of a BadgerDB.
type batchedMutations struct {
	kvStore          *badgerStore
	store            *badger.DB
	dbPrefix         []byte
	setOperations    map[string]kvstore.Value
	deleteOperations map[string]types.Empty
	operationsMutex  sync.Mutex
	closed           *atomic.Bool
}

func (b *batchedMutations) Set(key kvstore.Key, value kvstore.Value) error {
	stringKey := byteutils.ConcatBytesToString(b.dbPrefix, key)

	b.operationsMutex.Lock()
	defer b.operationsMutex.Unlock()

	delete(b.deleteOperations, stringKey)
	b.setOperations[stringKey] = value

	return nil
}

func (b *batchedMutations) Delete(key kvstore.Key) error {
	stringKey := byteutils.ConcatBytesToString(b.dbPrefix, key)

	b.operationsMutex.Lock()
	defer b.operationsMutex.Unlock()

	delete(b.setOperations, stringKey)
	b.deleteOperations[stringKey] = types.Void

	return nil
}

func (b *batchedMutations) Cancel() {
	b.operationsMutex.Lock()
	defer b.operationsMutex.Unlock()

	b.setOperations = make(map[string]kvstore.Value)
	b.deleteOperations = make(map[string]types.Empty)
}

func (b *batchedMutations) Commit() error {
	if b.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	writeBatch := b.store.NewWriteBatch()

	b.operationsMutex.Lock()
	defer b.operationsMutex.Unlock()

	for key, value := range b.setOperations {
		err := writeBatch.Set([]byte(key), value)
		if err != nil {
			return err
		}
	}

	for key := range b.deleteOperations {
		err := writeBatch.Delete([]byte(key))
		if err != nil {
			return err
		}
	}

	return writeBatch.Flush()
}

var (
	_ kvstore.KVStore          = &badgerStore{}
	_ kvstore.BatchedMutations = &batchedMutations{}
)

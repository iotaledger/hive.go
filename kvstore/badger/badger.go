package badger

import (
	"sync"

	"github.com/dgraph-io/badger/v2"
	"github.com/iotaledger/hive.go/types"
	"github.com/iotaledger/hive.go/typeutils"

	"github.com/iotaledger/hive.go/kvstore"
)

// KVStore implements the KVStore interface around a BadgerDB instance.
type badgerStore struct {
	instance *badger.DB
	dbPrefix []byte
}

// New creates a new KVStore with the underlying BadgerDB.
func New(db *badger.DB) kvstore.KVStore {
	return &badgerStore{
		instance: db,
	}
}

func (s *badgerStore) WithRealm(realm kvstore.Realm) kvstore.KVStore {
	return &badgerStore{
		instance: s.instance,
		dbPrefix: realm,
	}
}

func (s *badgerStore) Realm() []byte {
	return s.dbPrefix
}

// builds a key usable for the badger instance using the realm and the given prefix.
func (s *badgerStore) buildKeyPrefix(prefix kvstore.KeyPrefix) kvstore.KeyPrefix {
	value := s.dbPrefix
	return append(value, prefix...)
}

func (s *badgerStore) Iterate(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyValueConsumerFunc) error {
	return s.instance.View(func(txn *badger.Txn) (err error) {
		iteratorOptions := badger.DefaultIteratorOptions
		iteratorOptions.Prefix = s.buildKeyPrefix(prefix)
		iteratorOptions.PrefetchValues = true

		it := txn.NewIterator(iteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
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

func (s *badgerStore) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc) error {
	return s.instance.View(func(txn *badger.Txn) (err error) {
		iteratorOptions := badger.DefaultIteratorOptions
		iteratorOptions.Prefix = s.buildKeyPrefix(prefix)
		iteratorOptions.PrefetchValues = false

		it := txn.NewIterator(iteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			if !consumerFunc(it.Item().KeyCopy(nil)[len(s.dbPrefix):]) {
				break
			}
		}

		return nil
	})
}

func (s *badgerStore) Clear() error {
	return s.DeletePrefix(kvstore.EmptyPrefix)
}

func (s *badgerStore) Get(key kvstore.Key) (kvstore.Value, error) {
	var value []byte
	err := s.instance.View(func(txn *badger.Txn) error {
		item, err := txn.Get(append(s.dbPrefix, key...))
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)
		return err
	})
	if err == badger.ErrKeyNotFound {
		return nil, kvstore.ErrKeyNotFound
	}
	return value, nil
}

func (s *badgerStore) Set(key kvstore.Key, value kvstore.Value) error {
	return s.instance.Update(func(txn *badger.Txn) error {
		return txn.Set(append(s.dbPrefix, key...), value)
	})
}

func (s *badgerStore) Has(key kvstore.Key) (bool, error) {
	err := s.instance.View(func(txn *badger.Txn) error {
		_, err := txn.Get(append(s.dbPrefix, key...))
		return err
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *badgerStore) Delete(key kvstore.Key) error {
	err := s.instance.Update(func(txn *badger.Txn) error {
		return txn.Delete(append(s.dbPrefix, key...))
	})
	if err != nil && err == badger.ErrKeyNotFound {
		return kvstore.ErrKeyNotFound
	}
	return err
}

func (s *badgerStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
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

func (s *badgerStore) Batched() kvstore.BatchedMutations {
	return &batchedMutations{
		store:            s.instance,
		dbPrefix:         s.dbPrefix,
		setOperations:    make(map[string]kvstore.Value),
		deleteOperations: make(map[string]types.Empty),
	}
}

// batchedMutations is a wrapper around a WriteBatch of a BadgerDB.
type batchedMutations struct {
	store            *badger.DB
	dbPrefix         []byte
	setOperations    map[string]kvstore.Value
	deleteOperations map[string]types.Empty
	operationsMutex  sync.Mutex
}

func (b *batchedMutations) Set(key kvstore.Key, value kvstore.Value) error {
	stringKey := typeutils.BytesToString(append(b.dbPrefix, key...))

	b.operationsMutex.Lock()
	defer b.operationsMutex.Unlock()

	delete(b.deleteOperations, stringKey)
	b.setOperations[stringKey] = value

	return nil
}

func (b *batchedMutations) Delete(key kvstore.Key) error {
	stringKey := typeutils.BytesToString(append(b.dbPrefix, key...))

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
	writeBatch := b.store.NewWriteBatch()

	b.operationsMutex.Lock()
	defer b.operationsMutex.Unlock()

	for key, value := range b.setOperations {
		err := writeBatch.Set(typeutils.StringToBytes(key), value)
		if err != nil {
			return err
		}
	}

	for key, _ := range b.deleteOperations {
		err := writeBatch.Delete(typeutils.StringToBytes(key))
		if err != nil {
			return err
		}
	}

	return writeBatch.Flush()
}

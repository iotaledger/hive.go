package badger

import (
	"github.com/dgraph-io/badger/v2"

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
		batched:  s.instance.NewWriteBatch(),
		dbPrefix: s.dbPrefix,
	}
}

// batchedMutations is a wrapper around a WriteBatch of a BadgerDB.
type batchedMutations struct {
	batched  *badger.WriteBatch
	dbPrefix []byte
}

func (b *batchedMutations) Set(key kvstore.Key, value kvstore.Value) error {
	return b.batched.Set(append(b.dbPrefix, key...), value)
}

func (b *batchedMutations) Delete(key kvstore.Key) error {
	return b.batched.Delete(append(b.dbPrefix, key...))
}

func (b *batchedMutations) Cancel() {
	b.batched.Cancel()
}

func (b *batchedMutations) Commit() error {
	return b.batched.Flush()
}

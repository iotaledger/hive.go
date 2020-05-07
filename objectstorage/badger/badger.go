package badger

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/iotaledger/hive.go/objectstorage"
)

// New creates a new Storage with the underlying BadgerDB.
func New(db *badger.DB) *Storage {
	return &Storage{instance: db}
}

// Storage implements the ObjectStorage Storage interface around a BadgerDB instance.
type Storage struct {
	instance *badger.DB
}

// builds a key usable for the badger instance using the given realm and prefixes.
func buildPrefixedKey(realm []byte, prefixes [][]byte) []byte {
	prefix := realm
	for _, optionalPrefix := range prefixes {
		prefix = append(prefix, optionalPrefix...)
	}
	return prefix
}

func (s *Storage) Iterate(realm []byte, prefixes [][]byte, preFetchValues bool, consumerFunc objectstorage.IteratorKeyValueConsumerFunc) error {
	return s.instance.View(func(txn *badger.Txn) (err error) {
		iteratorOptions := badger.DefaultIteratorOptions
		iteratorOptions.Prefix = buildPrefixedKey(realm, prefixes)
		iteratorOptions.PrefetchValues = preFetchValues

		it := txn.NewIterator(iteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				panic(err)
			}
			if !consumerFunc(item.KeyCopy(nil)[len(realm):], value) {
				break
			}
		}

		return nil
	})
}

func (s *Storage) IterateKeys(realm []byte, prefixes [][]byte, consumerFunc objectstorage.IteratorKeyConsumerFunc) error {
	return s.instance.View(func(txn *badger.Txn) (err error) {
		iteratorOptions := badger.DefaultIteratorOptions
		iteratorOptions.Prefix = buildPrefixedKey(realm, prefixes)
		iteratorOptions.PrefetchValues = false

		it := txn.NewIterator(iteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			if !consumerFunc(it.Item().KeyCopy(nil)[len(realm):]) {
				break
			}
		}

		return nil
	})
}

func (s *Storage) Clear(realm []byte) error {
	return s.instance.DropPrefix(realm)
}

func (s *Storage) Get(realm []byte, key []byte) ([]byte, error) {
	var value []byte
	err := s.instance.View(func(txn *badger.Txn) error {
		item, err := txn.Get(append(realm, key...))
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)
		return err
	})
	if err == badger.ErrKeyNotFound {
		return nil, objectstorage.ErrKeyNotFound
	}
	return value, nil
}

func (s *Storage) Set(realm []byte, key []byte, value []byte) error {
	return s.instance.Update(func(txn *badger.Txn) error {
		return txn.Set(append(realm, key...), value)
	})
}

func (s *Storage) Has(realm []byte, key []byte) (bool, error) {
	err := s.instance.View(func(txn *badger.Txn) error {
		_, err := txn.Get(append(realm, key...))
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

func (s *Storage) Delete(realm []byte, key []byte) error {
	err := s.instance.Update(func(txn *badger.Txn) error {
		return txn.Delete(append(realm, key...))
	})
	if err != nil && err == badger.ErrKeyNotFound {
		return objectstorage.ErrKeyNotFound
	}
	return err
}

func (s *Storage) Batched() objectstorage.BatchedMutations {
	return &BatchedMutations{batched: s.instance.NewWriteBatch()}
}

// BatchedMutations is a wrapper around a WriteBatch of a BadgerDB.
type BatchedMutations struct {
	batched *badger.WriteBatch
}

func (batchedMuts *BatchedMutations) Set(realm []byte, key []byte, value []byte) error {
	return batchedMuts.batched.Set(append(realm, key...), value)
}

func (batchedMuts *BatchedMutations) Delete(realm []byte, key []byte) error {
	return batchedMuts.batched.Delete(append(realm, key...))
}

func (batchedMuts *BatchedMutations) Cancel() {
	batchedMuts.batched.Cancel()
}

func (batchedMuts *BatchedMutations) Commit() error {
	return batchedMuts.batched.Flush()
}

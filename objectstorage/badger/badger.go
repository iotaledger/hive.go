package badger

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/iotaledger/hive.go/objectstorage"
)

// Storage implements the ObjectStorage Storage interface around a BadgerDB instance.
type Storage struct {
	instance *badger.DB
	dbPrefix []byte
}

// New creates a new Storage with the underlying BadgerDB.
func New(db *badger.DB) objectstorage.Storage {
	return &Storage{
		instance: db,
	}
}

func (s *Storage) WithRealm(realm []byte) objectstorage.Storage {
	return &Storage{
		instance: s.instance,
		dbPrefix: realm,
	}
}

func (s *Storage) Realm() []byte {
	return s.dbPrefix
}

// builds a key usable for the badger instance using the realm and the given prefixes.
func (s *Storage) buildPrefixedKey(prefixes [][]byte) []byte {
	prefix := s.dbPrefix
	for _, optionalPrefix := range prefixes {
		prefix = append(prefix, optionalPrefix...)
	}
	return prefix
}

func (s *Storage) Iterate(prefixes [][]byte, preFetchValues bool, consumerFunc objectstorage.IteratorKeyValueConsumerFunc) error {
	return s.instance.View(func(txn *badger.Txn) (err error) {
		iteratorOptions := badger.DefaultIteratorOptions
		iteratorOptions.Prefix = s.buildPrefixedKey(prefixes)
		iteratorOptions.PrefetchValues = preFetchValues

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

func (s *Storage) IterateKeys(prefixes [][]byte, consumerFunc objectstorage.IteratorKeyConsumerFunc) error {
	return s.instance.View(func(txn *badger.Txn) (err error) {
		iteratorOptions := badger.DefaultIteratorOptions
		iteratorOptions.Prefix = s.buildPrefixedKey(prefixes)
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

func (s *Storage) Clear() error {
	return s.instance.DropPrefix(s.dbPrefix)
}

func (s *Storage) Get(key []byte) ([]byte, error) {
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
		return nil, objectstorage.ErrKeyNotFound
	}
	return value, nil
}

func (s *Storage) Set(key []byte, value []byte) error {
	return s.instance.Update(func(txn *badger.Txn) error {
		return txn.Set(append(s.dbPrefix, key...), value)
	})
}

func (s *Storage) Has(key []byte) (bool, error) {
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

func (s *Storage) Delete(key []byte) error {
	err := s.instance.Update(func(txn *badger.Txn) error {
		return txn.Delete(append(s.dbPrefix, key...))
	})
	if err != nil && err == badger.ErrKeyNotFound {
		return objectstorage.ErrKeyNotFound
	}
	return err
}

func (s *Storage) Batched() objectstorage.BatchedMutations {
	return &BatchedMutations{
		batched:  s.instance.NewWriteBatch(),
		dbPrefix: s.dbPrefix,
	}
}

// BatchedMutations is a wrapper around a WriteBatch of a BadgerDB.
type BatchedMutations struct {
	batched  *badger.WriteBatch
	dbPrefix []byte
}

func (b *BatchedMutations) Set(key []byte, value []byte) error {
	return b.batched.Set(append(b.dbPrefix, key...), value)
}

func (b *BatchedMutations) Delete(key []byte) error {
	return b.batched.Delete(append(b.dbPrefix, key...))
}

func (b *BatchedMutations) Cancel() {
	b.batched.Cancel()
}

func (b *BatchedMutations) Commit() error {
	return b.batched.Flush()
}

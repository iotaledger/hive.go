package badgerdb

import (
	"context"
	"errors"

	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/pb"

	"github.com/iotaledger/hive.go/database"
)

const (
	StreamNumGoRoutines = 16
)

type prefixedBadgerDB struct {
	db     *badger.DB
	prefix []byte
}

func NewDatabaseWithPrefix(prefix []byte, badgerInstance *badger.DB) database.Database {
	return &prefixedBadgerDB{
		db:     badgerInstance,
		prefix: prefix,
	}
}

func (pdb *prefixedBadgerDB) keyWithPrefix(key database.Key) database.Key {
	return append(pdb.prefix, key...)
}

func (pdb *prefixedBadgerDB) keyWithoutPrefix(key database.Key) database.Key {
	return key[1:]
}

func keyWithoutKeyPrefix(key database.Key, prefix database.KeyPrefix) database.Key {
	return key[len(prefix):]
}

func (pdb *prefixedBadgerDB) Set(entry database.Entry) error {
	wb := pdb.db.NewWriteBatch()
	defer wb.Cancel()

	e := badger.NewEntry(pdb.keyWithPrefix(entry.Key), entry.Value)
	err := wb.SetEntry(e)
	if err != nil {
		return err
	}
	return wb.Flush()
}

func (pdb *prefixedBadgerDB) Apply(set []database.Entry, delete []database.Key) error {

	wb := pdb.db.NewWriteBatch()
	defer wb.Cancel()

	for _, entry := range set {
		keyPrefix := pdb.keyWithPrefix(entry.Key)
		keyCopy := make([]byte, len(keyPrefix))
		copy(keyCopy, keyPrefix)

		valueCopy := make([]byte, len(entry.Value))
		copy(valueCopy, entry.Value)

		err := wb.SetEntry(badger.NewEntry(keyCopy, valueCopy))
		if err != nil {
			return err
		}
	}
	for _, key := range delete {
		keyPrefix := pdb.keyWithPrefix(key)
		keyCopy := make([]byte, len(keyPrefix))
		copy(keyCopy, keyPrefix)

		err := wb.Delete(keyCopy)
		if err != nil {
			return err
		}
	}
	return wb.Flush()
}

func (pdb *prefixedBadgerDB) Contains(key database.Key) (bool, error) {
	err := pdb.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(pdb.keyWithPrefix(key))
		return err
	})

	if err == badger.ErrKeyNotFound {
		return false, nil
	} else {
		return err == nil, err
	}
}

func (pdb *prefixedBadgerDB) Get(key database.Key) (database.Entry, error) {
	var result database.Entry

	err := pdb.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(pdb.keyWithPrefix(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return database.ErrKeyNotFound
			}
			return err
		}
		result.Key = key

		result.Value, err = item.ValueCopy(nil)
		return err
	})

	return result, err
}

func (pdb *prefixedBadgerDB) Delete(key database.Key) error {
	wb := pdb.db.NewWriteBatch()
	defer wb.Cancel()

	err := wb.Delete(pdb.keyWithPrefix(key))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return database.ErrKeyNotFound
		}
		return err
	}
	return wb.Flush()
}

func (pdb *prefixedBadgerDB) DeletePrefix(keyPrefix database.KeyPrefix) error {
	prefixToDelete := append(pdb.prefix, keyPrefix...)
	return pdb.db.DropPrefix(prefixToDelete)
}

// ForEach runs consumer for each valid DB Entry.
// Entry.Key is only valid as long as Entry is valid. If you need to modify it or use it outside, it must be copied.
func (pdb *prefixedBadgerDB) ForEach(consumer func(database.Entry) bool) error {
	err := pdb.db.View(func(txn *badger.Txn) error {
		iteratorOptions := badger.DefaultIteratorOptions
		it := txn.NewIterator(iteratorOptions)
		defer it.Close()
		prefix := pdb.prefix

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			if consumer(database.Entry{
				Key:   pdb.keyWithoutPrefix(item.Key()),
				Value: value,
			}) {
				break
			}
		}
		return nil
	})
	return err
}

// ForEachPrefix runs consumer for each valid DB entry matching keyPrefix.
// Entry.Key is only valid as long as Entry is valid. If you need to modify it or use it outside, it must be copied.
func (pdb *prefixedBadgerDB) ForEachPrefix(keyPrefix database.KeyPrefix, consumer func(database.Entry) bool) error {
	err := pdb.db.View(func(txn *badger.Txn) error {
		iteratorOptions := badger.DefaultIteratorOptions
		it := txn.NewIterator(iteratorOptions)
		defer it.Close()
		prefix := append(pdb.prefix, keyPrefix...)

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			if consumer(database.Entry{
				Key:   keyWithoutKeyPrefix(pdb.keyWithoutPrefix(item.Key()), keyPrefix),
				Value: value,
			}) {
				break
			}
		}
		return nil
	})
	return err
}

// ForEachPrefixKeyOnly runs consumer for each valid DB entry matching keyPrefix.
// KeyOnlyEntry.Key is only valid as long as KeyOnlyEntry is valid. If you need to modify it or use it outside, it must be copied.
func (pdb *prefixedBadgerDB) ForEachPrefixKeyOnly(keyPrefix database.KeyPrefix, consumer func(database.Key) bool) error {
	err := pdb.db.View(func(txn *badger.Txn) error {
		iteratorOptions := badger.DefaultIteratorOptions
		iteratorOptions.PrefetchValues = false
		it := txn.NewIterator(iteratorOptions)
		defer it.Close()
		prefix := append(pdb.prefix, keyPrefix...)

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			if consumer(keyWithoutKeyPrefix(pdb.keyWithoutPrefix(item.Key()), keyPrefix)) {
				break
			}
		}
		return nil
	})
	return err
}

func (pdb *prefixedBadgerDB) StreamForEach(consumer func(database.Entry) error) error {
	stream := pdb.db.NewStream()

	stream.NumGo = StreamNumGoRoutines
	stream.Prefix = pdb.prefix
	stream.ChooseKey = nil
	stream.KeyToList = nil

	// Send is called serially, while Stream.Orchestrate is running.
	stream.Send = func(list *pb.KVList) error {
		for _, kv := range list.Kv {
			err := consumer(database.Entry{
				Key:   pdb.keyWithoutPrefix(kv.GetKey()),
				Value: kv.GetValue(),
			})
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Run the stream
	return stream.Orchestrate(context.Background())
}

func (pdb *prefixedBadgerDB) StreamForEachKeyOnly(consumer func(database.Key) error) error {
	stream := pdb.db.NewStream()

	stream.NumGo = StreamNumGoRoutines
	stream.Prefix = pdb.prefix
	stream.ChooseKey = nil
	stream.KeyToList = nil

	// Send is called serially, while Stream.Orchestrate is running.
	stream.Send = func(list *pb.KVList) error {
		for _, kv := range list.Kv {
			err := consumer(pdb.keyWithoutPrefix(kv.GetKey()))
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Run the stream
	return stream.Orchestrate(context.Background())
}

func (pdb *prefixedBadgerDB) StreamForEachPrefix(keyPrefix database.KeyPrefix, consumer func(database.Entry) error) error {
	stream := pdb.db.NewStream()

	stream.NumGo = StreamNumGoRoutines
	stream.Prefix = append(pdb.prefix, keyPrefix...)
	stream.ChooseKey = nil
	stream.KeyToList = nil

	// Send is called serially, while Stream.Orchestrate is running.
	stream.Send = func(list *pb.KVList) error {
		for _, kv := range list.Kv {
			err := consumer(database.Entry{
				Key:   keyWithoutKeyPrefix(pdb.keyWithoutPrefix(kv.GetKey()), keyPrefix),
				Value: kv.GetValue(),
			})
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Run the stream
	return stream.Orchestrate(context.Background())
}

func (pdb *prefixedBadgerDB) StreamForEachPrefixKeyOnly(keyPrefix database.KeyPrefix, consumer func(database.Key) error) error {
	stream := pdb.db.NewStream()

	stream.NumGo = StreamNumGoRoutines
	stream.Prefix = append(pdb.prefix, keyPrefix...)
	stream.ChooseKey = nil
	stream.KeyToList = nil

	// Send is called serially, while Stream.Orchestrate is running.
	stream.Send = func(list *pb.KVList) error {
		for _, kv := range list.Kv {
			err := consumer(keyWithoutKeyPrefix(pdb.keyWithoutPrefix(kv.GetKey()), keyPrefix))
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Run the stream
	return stream.Orchestrate(context.Background())
}

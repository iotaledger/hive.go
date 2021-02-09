package badger

import (
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/types"
)

// badgerStore implements the KVStore interface around a BadgerDB instance.
type badgerStore struct {
	instance                     *badger.DB
	dbPrefix                     []byte
	accessCallback               kvstore.AccessCallback
	accessCallbackCommandsFilter kvstore.Command
}

// New creates a new KVStore with the underlying BadgerDB.
func New(db *badger.DB) kvstore.KVStore {
	return &badgerStore{
		instance: db,
	}
}

// AccessCallback configures the store to pass all requests to the KVStore to the given callback.
// This can for example be used for debugging and to examine what the KVStore is doing.
func (s *badgerStore) AccessCallback(callback kvstore.AccessCallback, commandsFilter ...kvstore.Command) {
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
	return byteutils.ConcatBytes(s.dbPrefix, prefix)
}

// Shutdown marks the store as shutdown.
func (s *badgerStore) Shutdown() {
	if s.accessCallback != nil {
		s.accessCallback(kvstore.ShutdownCommand)
	}
}

func (s *badgerStore) Iterate(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyValueConsumerFunc) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.IterateCommand) {
		s.accessCallback(kvstore.IterateCommand, prefix)
	}

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
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.IterateKeysCommand) {
		s.accessCallback(kvstore.IterateKeysCommand, prefix)
	}

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
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.ClearCommand) {
		s.accessCallback(kvstore.ClearCommand)
	}

	return s.DeletePrefix(kvstore.EmptyPrefix)
}

func (s *badgerStore) Get(key kvstore.Key) (kvstore.Value, error) {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.GetCommand) {
		s.accessCallback(kvstore.GetCommand, key)
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
	if err == badger.ErrKeyNotFound {
		return nil, kvstore.ErrKeyNotFound
	}

	return value, nil
}

func (s *badgerStore) Set(key kvstore.Key, value kvstore.Value) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.SetCommand) {
		s.accessCallback(kvstore.SetCommand, key, value)
	}

	return s.instance.Update(func(txn *badger.Txn) error {
		return txn.Set(byteutils.ConcatBytes(s.dbPrefix, key), value)
	})
}

func (s *badgerStore) Has(key kvstore.Key) (bool, error) {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.HasCommand) {
		s.accessCallback(kvstore.HasCommand, key)
	}

	err := s.instance.View(func(txn *badger.Txn) error {
		_, err := txn.Get(byteutils.ConcatBytes(s.dbPrefix, key))
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
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.DeleteCommand) {
		s.accessCallback(kvstore.DeleteCommand, key)
	}

	err := s.instance.Update(func(txn *badger.Txn) error {
		return txn.Delete(byteutils.ConcatBytes(s.dbPrefix, key))
	})
	if err != nil && err == badger.ErrKeyNotFound {
		return kvstore.ErrKeyNotFound
	}
	return err
}

func (s *badgerStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.DeletePrefixCommand) {
		s.accessCallback(kvstore.DeletePrefixCommand, prefix)
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

func (s *badgerStore) Batched() kvstore.BatchedMutations {
	return &batchedMutations{
		kvStore:          s,
		store:            s.instance,
		dbPrefix:         s.dbPrefix,
		setOperations:    make(map[string]kvstore.Value),
		deleteOperations: make(map[string]types.Empty),
	}
}

func (s *badgerStore) Flush() error {
	return s.instance.Sync()
}

func (s *badgerStore) Close() error {
	return s.instance.Close()
}

// batchedMutations is a wrapper around a WriteBatch of a BadgerDB.
type batchedMutations struct {
	kvStore          *badgerStore
	store            *badger.DB
	dbPrefix         []byte
	setOperations    map[string]kvstore.Value
	deleteOperations map[string]types.Empty
	operationsMutex  sync.Mutex
}

func (b *batchedMutations) Set(key kvstore.Key, value kvstore.Value) error {
	if b.kvStore.accessCallback != nil && b.kvStore.accessCallbackCommandsFilter.HasBits(kvstore.SetCommand) {
		b.kvStore.accessCallback(kvstore.SetCommand, key, value)
	}

	stringKey := byteutils.ConcatBytesToString(b.dbPrefix, key)

	b.operationsMutex.Lock()
	defer b.operationsMutex.Unlock()

	delete(b.deleteOperations, stringKey)
	b.setOperations[stringKey] = value

	return nil
}

func (b *batchedMutations) Delete(key kvstore.Key) error {
	if b.kvStore.accessCallback != nil && b.kvStore.accessCallbackCommandsFilter.HasBits(kvstore.DeleteCommand) {
		b.kvStore.accessCallback(kvstore.DeleteCommand, key)
	}

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

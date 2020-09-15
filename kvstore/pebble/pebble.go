package pebble

import (
	"sync"

	"github.com/cockroachdb/pebble"

	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/types"
)

// peppleStore implements the KVStore interface around a pebble instance.
type peppleStore struct {
	instance                     *pebble.DB
	dbPrefix                     []byte
	accessCallback               kvstore.AccessCallback
	accessCallbackCommandsFilter kvstore.Command
}

// New creates a new KVStore with the underlying pebbleDB.
func New(db *pebble.DB) kvstore.KVStore {
	return &peppleStore{
		instance: db,
	}
}

// AccessCallback configures the store to pass all requests to the KVStore to the given callback.
// This can for example be used for debugging and to examine what the KVStore is doing.
func (s *peppleStore) AccessCallback(callback kvstore.AccessCallback, commandsFilter ...kvstore.Command) {
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

func (s *peppleStore) WithRealm(realm kvstore.Realm) kvstore.KVStore {
	return &peppleStore{
		instance: s.instance,
		dbPrefix: realm,
	}
}

func (s *peppleStore) Realm() []byte {
	return s.dbPrefix
}

// builds a key usable for the pebble instance using the realm and the given prefix.
func (s *peppleStore) buildKeyPrefix(prefix kvstore.KeyPrefix) kvstore.KeyPrefix {
	return byteutils.ConcatBytes(s.dbPrefix, prefix)
}

// Shutdown marks the store as shutdown.
func (s *peppleStore) Shutdown() {
	if s.accessCallback != nil {
		s.accessCallback(kvstore.ShutdownCommand)
	}
}

func (s *peppleStore) Iterate(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyValueConsumerFunc) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.IterateCommand) {
		s.accessCallback(kvstore.IterateCommand, prefix)
	}

	start := s.buildKeyPrefix(prefix)
	end := copyBytes(start)
	end[len(end)-1] = end[len(end)-1] + 1

	iter := s.instance.NewIter(&pebble.IterOptions{LowerBound: start, UpperBound: end})
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		if !consumerFunc(copyBytes(iter.Key())[len(s.dbPrefix):], copyBytes(iter.Value())) {
			break
		}
	}

	return nil
}

func (s *peppleStore) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.IterateKeysCommand) {
		s.accessCallback(kvstore.IterateKeysCommand, prefix)
	}

	start := s.buildKeyPrefix(prefix)
	end := copyBytes(start)
	end[len(end)-1] = end[len(end)-1] + 1

	iter := s.instance.NewIter(&pebble.IterOptions{LowerBound: start, UpperBound: end})
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		if !consumerFunc(copyBytes(iter.Key())[len(s.dbPrefix):]) {
			break
		}
	}

	return nil
}

func (s *peppleStore) Clear() error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.ClearCommand) {
		s.accessCallback(kvstore.ClearCommand)
	}

	return s.DeletePrefix(kvstore.EmptyPrefix)
}

func (s *peppleStore) Get(key kvstore.Key) (kvstore.Value, error) {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.GetCommand) {
		s.accessCallback(kvstore.GetCommand, key)
	}

	val, closer, err := s.instance.Get(byteutils.ConcatBytes(s.dbPrefix, key))

	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, kvstore.ErrKeyNotFound
		}
		return nil, err
	}

	value := copyBytes(val)

	if err := closer.Close(); err != nil {
		return nil, err
	}

	return value, nil
}

func (s *peppleStore) Set(key kvstore.Key, value kvstore.Value) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.SetCommand) {
		s.accessCallback(kvstore.SetCommand, key, value)
	}

	return s.instance.Set(byteutils.ConcatBytes(s.dbPrefix, key), value, pebble.NoSync)
}

func (s *peppleStore) Has(key kvstore.Key) (bool, error) {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.HasCommand) {
		s.accessCallback(kvstore.HasCommand, key)
	}

	_, closer, err := s.instance.Get(byteutils.ConcatBytes(s.dbPrefix, key))
	if err == pebble.ErrNotFound {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if err := closer.Close(); err != nil {
		return true, err
	}

	return true, nil
}

func (s *peppleStore) Delete(key kvstore.Key) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.DeleteCommand) {
		s.accessCallback(kvstore.DeleteCommand, key)
	}

	return s.instance.Delete(byteutils.ConcatBytes(s.dbPrefix, key), pebble.NoSync)
}

func (s *peppleStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(kvstore.DeletePrefixCommand) {
		s.accessCallback(kvstore.DeletePrefixCommand, prefix)
	}

	start := s.buildKeyPrefix(prefix)
	end := copyBytes(start)
	end[len(end)-1] = end[len(end)-1] + 1

	return s.instance.DeleteRange(start, end, pebble.NoSync)
}

func (s *peppleStore) Batched() kvstore.BatchedMutations {
	return &batchedMutations{
		kvStore:          s,
		store:            s.instance,
		dbPrefix:         s.dbPrefix,
		setOperations:    make(map[string]kvstore.Value),
		deleteOperations: make(map[string]types.Empty),
	}
}

// batchedMutations is a wrapper around a WriteBatch of a pebbleDB.
type batchedMutations struct {
	kvStore          *peppleStore
	store            *pebble.DB
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
	writeBatch := b.store.NewBatch()

	b.operationsMutex.Lock()
	defer b.operationsMutex.Unlock()

	for key, value := range b.setOperations {
		err := writeBatch.Set([]byte(key), value, nil)
		if err != nil {
			return err
		}
	}

	for key := range b.deleteOperations {
		err := writeBatch.Delete([]byte(key), nil)
		if err != nil {
			return err
		}
	}

	return writeBatch.Commit(pebble.NoSync)
}

package pebble

import (
	"sync"

	"github.com/cockroachdb/pebble"

	"github.com/iotaledger/hive.go/v2/byteutils"
	"github.com/iotaledger/hive.go/v2/kvstore"
	"github.com/iotaledger/hive.go/v2/kvstore/utils"
	"github.com/iotaledger/hive.go/v2/types"
)

// pebbleStore implements the KVStore interface around a pebble instance.
type pebbleStore struct {
	instance *pebble.DB
	dbPrefix []byte
}

// New creates a new KVStore with the underlying pebbleDB.
func New(db *pebble.DB) kvstore.KVStore {
	return &pebbleStore{
		instance: db,
	}
}

func (s *pebbleStore) WithRealm(realm kvstore.Realm) kvstore.KVStore {
	return &pebbleStore{
		instance: s.instance,
		dbPrefix: realm,
	}
}

func (s *pebbleStore) Realm() []byte {
	return s.dbPrefix
}

// builds a key usable for the pebble instance using the realm and the given prefix.
func (s *pebbleStore) buildKeyPrefix(prefix kvstore.KeyPrefix) kvstore.KeyPrefix {
	return byteutils.ConcatBytes(s.dbPrefix, prefix)
}

// Shutdown marks the store as shutdown.
func (s *pebbleStore) Shutdown() {
}

func (s *pebbleStore) getIterBounds(prefix []byte) ([]byte, []byte) {
	start := s.buildKeyPrefix(prefix)

	if len(start) == 0 {
		// no bounds
		return nil, nil
	}

	return start, utils.KeyPrefixUpperBound(start)
}

// getIterFuncs returns the function pointers for the iteration based on the given settings.
func (s *pebbleStore) getIterFuncs(it *pebble.Iterator, iterDirection ...kvstore.IterDirection) (start func() bool, valid func() bool, move func() bool, err error) {

	startFunc := it.First
	validFunc := it.Valid
	moveFunc := it.Next

	if kvstore.GetIterDirection(iterDirection...) == kvstore.IterDirectionBackward {
		startFunc = it.Last
		moveFunc = it.Prev
	}

	return startFunc, validFunc, moveFunc, nil
}

// Iterate iterates over all keys and values with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys and values.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (s *pebbleStore) Iterate(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyValueConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	start, end := s.getIterBounds(prefix)

	it := s.instance.NewIter(&pebble.IterOptions{LowerBound: start, UpperBound: end})
	defer it.Close()

	startFunc, validFunc, moveFunc, err := s.getIterFuncs(it, iterDirection...)
	if err != nil {
		return err
	}

	for startFunc(); validFunc(); moveFunc() {
		if !consumerFunc(utils.CopyBytes(it.Key())[len(s.dbPrefix):], utils.CopyBytes(it.Value())) {
			break
		}
	}

	return nil
}

// IterateKeys iterates over all keys with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (s *pebbleStore) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	start, end := s.getIterBounds(prefix)

	it := s.instance.NewIter(&pebble.IterOptions{LowerBound: start, UpperBound: end})
	defer it.Close()

	startFunc, validFunc, moveFunc, err := s.getIterFuncs(it, iterDirection...)
	if err != nil {
		return err
	}

	for startFunc(); validFunc(); moveFunc() {
		if !consumerFunc(utils.CopyBytes(it.Key())[len(s.dbPrefix):]) {
			break
		}
	}

	return nil
}

func (s *pebbleStore) Clear() error {
	return s.DeletePrefix(kvstore.EmptyPrefix)
}

func (s *pebbleStore) Get(key kvstore.Key) (kvstore.Value, error) {
	val, closer, err := s.instance.Get(byteutils.ConcatBytes(s.dbPrefix, key))

	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, kvstore.ErrKeyNotFound
		}
		return nil, err
	}

	value := utils.CopyBytes(val)

	if err := closer.Close(); err != nil {
		return nil, err
	}

	return value, nil
}

func (s *pebbleStore) Set(key kvstore.Key, value kvstore.Value) error {
	return s.instance.Set(byteutils.ConcatBytes(s.dbPrefix, key), value, pebble.NoSync)
}

func (s *pebbleStore) Has(key kvstore.Key) (bool, error) {
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

func (s *pebbleStore) Delete(key kvstore.Key) error {
	return s.instance.Delete(byteutils.ConcatBytes(s.dbPrefix, key), pebble.NoSync)
}

func (s *pebbleStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	start, end := s.getIterBounds(prefix)

	if start == nil {
		// DeleteRange does not work without range, so we have to iterate over all keys and delete them
		it := s.instance.NewIter(&pebble.IterOptions{LowerBound: start, UpperBound: end})
		defer it.Close()

		b := s.instance.NewBatch()
		for it.First(); it.Valid(); it.Next() {
			if err := b.Delete(it.Key(), nil); err != nil {
				b.Close()
				return err
			}
		}

		return b.Commit(pebble.NoSync)
	}

	return s.instance.DeleteRange(start, end, pebble.NoSync)
}

func (s *pebbleStore) Batched() kvstore.BatchedMutations {
	return &batchedMutations{
		kvStore:          s,
		store:            s.instance,
		dbPrefix:         s.dbPrefix,
		setOperations:    make(map[string]kvstore.Value),
		deleteOperations: make(map[string]types.Empty),
	}
}

func (s *pebbleStore) Flush() error {
	return s.instance.Flush()
}

func (s *pebbleStore) Close() error {
	return s.instance.Close()
}

// batchedMutations is a wrapper around a WriteBatch of a pebbleDB.
type batchedMutations struct {
	kvStore          *pebbleStore
	store            *pebble.DB
	dbPrefix         []byte
	setOperations    map[string]kvstore.Value
	deleteOperations map[string]types.Empty
	operationsMutex  sync.Mutex
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

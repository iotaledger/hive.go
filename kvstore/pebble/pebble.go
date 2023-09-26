package pebble

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/cockroachdb/pebble"

	"github.com/izuc/zipp.foundation/ds/types"
	"github.com/izuc/zipp.foundation/kvstore"
	"github.com/izuc/zipp.foundation/kvstore/utils"
	"github.com/izuc/zipp.foundation/serializer/v2/byteutils"
)

// pebbleStore implements the KVStore interface around a pebble instance.
type pebbleStore struct {
	instance *pebble.DB
	closed   *atomic.Bool
	dbPrefix []byte
}

// New creates a new KVStore with the underlying pebbleDB.
func New(db *pebble.DB) kvstore.KVStore {
	return &pebbleStore{
		instance: db,
		closed:   new(atomic.Bool),
	}
}

func (s *pebbleStore) WithRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	if s.closed.Load() {
		return nil, kvstore.ErrStoreClosed
	}

	return &pebbleStore{
		instance: s.instance,
		closed:   s.closed,
		dbPrefix: realm,
	}, nil
}

func (s *pebbleStore) WithExtendedRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	return s.WithRealm(byteutils.ConcatBytes(s.Realm(), realm))
}

func (s *pebbleStore) Realm() []byte {
	return s.dbPrefix
}

// builds a key usable for the pebble instance using the realm and the given prefix.
func (s *pebbleStore) buildKeyPrefix(prefix kvstore.KeyPrefix) kvstore.KeyPrefix {
	return byteutils.ConcatBytes(s.dbPrefix, prefix)
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
func (s *pebbleStore) getIterFuncs(it *pebble.Iterator, iterDirection ...kvstore.IterDirection) (start func() bool, valid func() bool, move func() bool) {

	startFunc := it.First
	validFunc := it.Valid
	moveFunc := it.Next

	if kvstore.GetIterDirection(iterDirection...) == kvstore.IterDirectionBackward {
		startFunc = it.Last
		moveFunc = it.Prev
	}

	return startFunc, validFunc, moveFunc
}

// Iterate iterates over all keys and values with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys and values.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (s *pebbleStore) Iterate(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyValueConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	start, end := s.getIterBounds(prefix)

	it := s.instance.NewIter(&pebble.IterOptions{LowerBound: start, UpperBound: end})
	defer it.Close()

	startFunc, validFunc, moveFunc := s.getIterFuncs(it, iterDirection...)

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
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	start, end := s.getIterBounds(prefix)

	it := s.instance.NewIter(&pebble.IterOptions{LowerBound: start, UpperBound: end})
	defer it.Close()

	startFunc, validFunc, moveFunc := s.getIterFuncs(it, iterDirection...)

	for startFunc(); validFunc(); moveFunc() {
		if !consumerFunc(utils.CopyBytes(it.Key())[len(s.dbPrefix):]) {
			break
		}
	}

	return nil
}

func (s *pebbleStore) Clear() error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	return s.DeletePrefix(kvstore.EmptyPrefix)
}

func (s *pebbleStore) Get(key kvstore.Key) (kvstore.Value, error) {
	if s.closed.Load() {
		return nil, kvstore.ErrStoreClosed
	}

	val, closer, err := s.instance.Get(byteutils.ConcatBytes(s.dbPrefix, key))

	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
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
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	return s.instance.Set(byteutils.ConcatBytes(s.dbPrefix, key), value, pebble.NoSync)
}

func (s *pebbleStore) Has(key kvstore.Key) (bool, error) {
	if s.closed.Load() {
		return false, kvstore.ErrStoreClosed
	}

	_, closer, err := s.instance.Get(byteutils.ConcatBytes(s.dbPrefix, key))
	if errors.Is(err, pebble.ErrNotFound) {
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
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	return s.instance.Delete(byteutils.ConcatBytes(s.dbPrefix, key), pebble.NoSync)
}

func (s *pebbleStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

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

func (s *pebbleStore) Flush() error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	return s.instance.Flush()
}

func (s *pebbleStore) Close() error {
	if s.closed.Swap(true) {
		// was already closed
		return nil
	}

	return s.instance.Close()
}

func (s *pebbleStore) Batched() (kvstore.BatchedMutations, error) {
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

// batchedMutations is a wrapper around a WriteBatch of a pebbleDB.
type batchedMutations struct {
	kvStore          *pebbleStore
	store            *pebble.DB
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

var _ kvstore.KVStore = &pebbleStore{}
var _ kvstore.BatchedMutations = &batchedMutations{}

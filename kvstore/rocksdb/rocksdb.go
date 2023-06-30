//go:build rocksdb

package rocksdb

import (
	"sync"
	"sync/atomic"

	"github.com/iotaledger/grocksdb"
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/utils"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
)

type rocksDBStore struct {
	instance *RocksDB
	dbPrefix []byte
	closed   *atomic.Bool
}

// New creates a new KVStore with the underlying RocksDB.
func New(db *RocksDB) kvstore.KVStore {
	return &rocksDBStore{
		instance: db,
		closed:   new(atomic.Bool),
	}
}

func (s *rocksDBStore) WithRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	if s.closed.Load() {
		return nil, kvstore.ErrStoreClosed
	}

	return &rocksDBStore{
		instance: s.instance,
		closed:   s.closed,
		dbPrefix: realm,
	}, nil
}

func (s *rocksDBStore) WithExtendedRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	return s.WithRealm(byteutils.ConcatBytes(s.Realm(), realm))
}

func (s *rocksDBStore) Realm() []byte {
	return s.dbPrefix
}

// builds a key usable using the realm and the given prefix.
func (s *rocksDBStore) buildKeyPrefix(prefix kvstore.KeyPrefix) kvstore.KeyPrefix {
	return byteutils.ConcatBytes(s.dbPrefix, prefix)
}

// getIterFuncs returns the function pointers for the iteration based on the given settings.
func (s *rocksDBStore) getIterFuncs(it *grocksdb.Iterator, keyPrefix []byte, iterDirection ...kvstore.IterDirection) (start func(), valid func() bool, move func(), err error) {

	startFunc := it.SeekToFirst
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
		startFunc = it.SeekToLast
		moveFunc = it.Prev

		if len(keyPrefix) > 0 {
			// we need to search the first item after the prefix
			prefixUpperBound := utils.KeyPrefixUpperBound(keyPrefix)
			if prefixUpperBound == nil {
				return nil, nil, nil, ierrors.New("no upper bound for prefix")
			}
			startFunc = func() {
				it.SeekForPrev(prefixUpperBound)

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
func (s *rocksDBStore) Iterate(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyValueConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	it := s.instance.db.NewIterator(s.instance.ro)
	defer it.Close()

	startFunc, validFunc, moveFunc, err := s.getIterFuncs(it, s.buildKeyPrefix(prefix), iterDirection...)
	if err != nil {
		return err
	}

	for startFunc(); validFunc(); moveFunc() {
		key := it.Key()
		k := utils.CopyBytes(key.Data(), key.Size())[len(s.dbPrefix):]
		key.Free()

		value := it.Value()
		v := utils.CopyBytes(value.Data(), value.Size())
		value.Free()

		if !consumerFunc(k, v) {
			break
		}
	}

	return nil
}

// IterateKeys iterates over all keys with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (s *rocksDBStore) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	it := s.instance.db.NewIterator(s.instance.ro)
	defer it.Close()

	startFunc, validFunc, moveFunc, err := s.getIterFuncs(it, s.buildKeyPrefix(prefix), iterDirection...)
	if err != nil {
		return err
	}

	for startFunc(); validFunc(); moveFunc() {
		key := it.Key()
		k := utils.CopyBytes(key.Data(), key.Size())[len(s.dbPrefix):]
		key.Free()

		if !consumerFunc(k) {
			break
		}
	}

	return nil
}

func (s *rocksDBStore) Clear() error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	return s.DeletePrefix(kvstore.EmptyPrefix)
}

func (s *rocksDBStore) Get(key kvstore.Key) (kvstore.Value, error) {
	if s.closed.Load() {
		return nil, kvstore.ErrStoreClosed
	}

	v, err := s.instance.db.GetBytes(s.instance.ro, byteutils.ConcatBytes(s.dbPrefix, key))
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, kvstore.ErrKeyNotFound
	}
	return v, nil
}

func (s *rocksDBStore) Set(key kvstore.Key, value kvstore.Value) error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	return s.instance.db.Put(s.instance.wo, byteutils.ConcatBytes(s.dbPrefix, key), value)
}

func (s *rocksDBStore) Has(key kvstore.Key) (bool, error) {
	if s.closed.Load() {
		return false, kvstore.ErrStoreClosed
	}

	v, err := s.instance.db.Get(s.instance.ro, byteutils.ConcatBytes(s.dbPrefix, key))
	defer v.Free()
	if err != nil {
		return false, err
	}
	return v.Exists(), nil
}

func (s *rocksDBStore) Delete(key kvstore.Key) error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	return s.instance.db.Delete(s.instance.wo, byteutils.ConcatBytes(s.dbPrefix, key))
}

func (s *rocksDBStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	keyPrefix := s.buildKeyPrefix(prefix)

	writeBatch := grocksdb.NewWriteBatch()
	defer writeBatch.Destroy()

	it := s.instance.db.NewIterator(s.instance.ro)
	defer it.Close()

	for it.Seek(keyPrefix); it.ValidForPrefix(keyPrefix); it.Next() {
		key := it.Key()
		writeBatch.Delete(key.Data())
		key.Free()
	}

	return s.instance.db.Write(s.instance.wo, writeBatch)
}

func (s *rocksDBStore) Flush() error {
	if s.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	return s.instance.Flush()
}

func (s *rocksDBStore) Close() error {
	if s.closed.Swap(true) {
		// was already closed
		return nil
	}

	return s.instance.Close()
}

func (s *rocksDBStore) Batched() (kvstore.BatchedMutations, error) {
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

// batchedMutations is a wrapper around a WriteBatch of a rocksDB.
type batchedMutations struct {
	kvStore          *rocksDBStore
	store            *RocksDB
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

	writeBatch := grocksdb.NewWriteBatch()
	defer writeBatch.Destroy()

	b.operationsMutex.Lock()
	defer b.operationsMutex.Unlock()

	for key, value := range b.setOperations {
		writeBatch.Put([]byte(key), value)
	}

	for key := range b.deleteOperations {
		writeBatch.Delete([]byte(key))
	}

	return b.store.db.Write(b.store.wo, writeBatch)
}

var _ kvstore.KVStore = &rocksDBStore{}
var _ kvstore.BatchedMutations = &batchedMutations{}

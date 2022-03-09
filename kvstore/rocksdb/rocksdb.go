//go:build rocksdb
// +build rocksdb

package rocksdb

import (
	"errors"
	"sync"

	"github.com/linxGnu/grocksdb"

	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/utils"
	"github.com/iotaledger/hive.go/types"
)

type rocksDBStore struct {
	instance *RocksDB
	dbPrefix []byte
}

// New creates a new KVStore with the underlying RocksDB.
func New(db *RocksDB) kvstore.KVStore {
	return &rocksDBStore{
		instance: db,
	}
}

func (s *rocksDBStore) WithRealm(realm kvstore.Realm) kvstore.KVStore {
	return &rocksDBStore{
		instance: s.instance,
		dbPrefix: realm,
	}
}

func (s *rocksDBStore) Realm() []byte {
	return s.dbPrefix
}

// builds a key usable using the realm and the given prefix.
func (s *rocksDBStore) buildKeyPrefix(prefix kvstore.KeyPrefix) kvstore.KeyPrefix {
	return byteutils.ConcatBytes(s.dbPrefix, prefix)
}

// Shutdown marks the store as shutdown.
func (s *rocksDBStore) Shutdown() {
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
				return nil, nil, nil, errors.New("no upper bound for prefix")
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
	return s.DeletePrefix(kvstore.EmptyPrefix)
}

func (s *rocksDBStore) Get(key kvstore.Key) (kvstore.Value, error) {
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
	return s.instance.db.Put(s.instance.wo, byteutils.ConcatBytes(s.dbPrefix, key), value)
}

func (s *rocksDBStore) Has(key kvstore.Key) (bool, error) {
	v, err := s.instance.db.Get(s.instance.ro, byteutils.ConcatBytes(s.dbPrefix, key))
	defer v.Free()
	if err != nil {
		return false, err
	}
	return v.Exists(), nil
}

func (s *rocksDBStore) Delete(key kvstore.Key) error {
	return s.instance.db.Delete(s.instance.wo, byteutils.ConcatBytes(s.dbPrefix, key))
}

func (s *rocksDBStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
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

func (s *rocksDBStore) Batched() kvstore.BatchedMutations {
	return &batchedMutations{
		kvStore:          s,
		store:            s.instance,
		dbPrefix:         s.dbPrefix,
		setOperations:    make(map[string]kvstore.Value),
		deleteOperations: make(map[string]types.Empty),
	}
}

func (s *rocksDBStore) Flush() error {
	return s.instance.Flush()
}

func (s *rocksDBStore) Close() error {
	return s.instance.Close()
}

// batchedMutations is a wrapper around a WriteBatch of a rocksDB.
type batchedMutations struct {
	kvStore          *rocksDBStore
	store            *RocksDB
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

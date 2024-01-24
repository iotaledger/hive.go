package ads

import (
	"crypto/sha256"
	"sync"

	"github.com/pokt-network/smt"

	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2/typeutils"
)

const (
	prefixRawKeysStorage uint8 = iota
	prefixTreeStorage
	prefixRootKey
	prefixSizeKey
)

// AuthenticatedMap is a sparse merkle tree based map.
type authenticatedMap[IdentifierType types.IdentifierType, K, V any] struct {
	rawKeysStore *kvstore.TypedStore[K, types.Empty]
	tree         *smt.SMT
	size         *kvstore.TypedValue[uint64]
	root         *kvstore.TypedValue[IdentifierType]
	mutex        sync.RWMutex

	keyToBytes   kvstore.ObjectToBytes[K]
	valueToBytes kvstore.ObjectToBytes[V]
	bytesToValue kvstore.BytesToObject[V]
}

// NewAuthenticatedMap creates a new authenticated map.
func newAuthenticatedMap[IdentifierType types.IdentifierType, K, V any](
	store kvstore.KVStore,
	identifierToBytes kvstore.ObjectToBytes[IdentifierType],
	bytesToIdentifier kvstore.BytesToObject[IdentifierType],
	keyToBytes kvstore.ObjectToBytes[K],
	bytesToKey kvstore.BytesToObject[K],
	valueToBytes kvstore.ObjectToBytes[V],
	bytesToValue kvstore.BytesToObject[V],
) *authenticatedMap[IdentifierType, K, V] {
	newMap := &authenticatedMap[IdentifierType, K, V]{
		rawKeysStore: kvstore.NewTypedStore(lo.PanicOnErr(store.WithExtendedRealm([]byte{prefixRawKeysStorage})), keyToBytes, bytesToKey, types.Empty.Bytes, types.EmptyFromBytes),
		size:         kvstore.NewTypedValue(store, []byte{prefixSizeKey}, typeutils.Uint64ToBytes, typeutils.Uint64FromBytes),
		root:         kvstore.NewTypedValue(store, []byte{prefixRootKey}, identifierToBytes, bytesToIdentifier),

		keyToBytes:   keyToBytes,
		valueToBytes: valueToBytes,
		bytesToValue: bytesToValue,
	}

	mapStoreAdapter := newMapStoreAdapter(lo.PanicOnErr(store.WithExtendedRealm([]byte{prefixTreeStorage})))
	if root, err := newMap.root.Get(); err == nil {
		newMap.tree = smt.ImportSparseMerkleTrie(mapStoreAdapter, sha256.New(), root[:], smt.WithValueHasher(nil))
	} else {
		newMap.tree = smt.NewSparseMerkleTrie(mapStoreAdapter, sha256.New(), smt.WithValueHasher(nil))
	}

	return newMap
}

// WasRestoredFromStorage returns true if the map has been restored from storage.
func (m *authenticatedMap[IdentifierType, K, V]) WasRestoredFromStorage() bool {
	_, err := m.root.Get()
	return !ierrors.Is(err, kvstore.ErrKeyNotFound)
}

// Root returns the root of the state sparse merkle tree at the latest committed slot.
func (m *authenticatedMap[IdentifierType, K, V]) Root() (root IdentifierType) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return IdentifierType(m.tree.Root())
}

// Set sets the output to unspent outputs set.
func (m *authenticatedMap[IdentifierType, K, V]) Set(key K, value V) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	valueBytes, err := m.valueToBytes(value)
	if err != nil {
		return ierrors.Wrap(err, "failed to serialize value")
	}

	keyBytes, err := m.keyToBytes(key)
	if err != nil {
		return ierrors.Wrap(err, "failed to serialize key")
	}

	has, err := m.has(keyBytes)
	if err != nil {
		return ierrors.Wrap(err, "failed to check if key exists")
	}

	if err := m.tree.Update(keyBytes, valueBytes); err != nil {
		return ierrors.Wrap(err, "failed to update tree")
	}

	if err := m.rawKeysStore.Set(key, types.Void); err != nil {
		return ierrors.Wrap(err, "failed to set raw key")
	}

	if !has {
		if err := m.addSize(1); err != nil {
			return ierrors.Wrap(err, "failed to increase size")
		}
	}

	return nil
}

// Size returns the number of elements in the map.
func (m *authenticatedMap[IdentifierType, K, V]) Size() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	size, err := m.size.Get()
	if err != nil {
		return 0
	}

	return int(size)
}

// Commit persists the current state of the map to the storage.
func (m *authenticatedMap[IdentifierType, K, V]) Commit() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if err := m.root.Set(IdentifierType(m.tree.Root())); err != nil {
		return ierrors.Wrap(err, "failed to set root")
	}

	return m.tree.Commit()
}

// Delete removes the key from the map.
func (m *authenticatedMap[IdentifierType, K, V]) Delete(key K) (deleted bool, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	keyBytes, err := m.keyToBytes(key)
	if err != nil {
		return false, ierrors.Wrap(err, "failed to serialize key")
	}

	has, err := m.has(keyBytes)
	if err != nil {
		return false, ierrors.Wrap(err, "failed to check if key exists")
	}

	if !has {
		return false, nil
	}

	if err := m.tree.Delete(keyBytes); err != nil {
		return false, ierrors.Wrap(err, "failed to delete from tree")
	}

	if err := m.rawKeysStore.Delete(key); err != nil {
		return false, ierrors.Wrap(err, "failed to delete from raw keys store")
	}

	if has {
		if err := m.addSize(-1); err != nil {
			return false, ierrors.Wrap(err, "failed to decrease size")
		}
	}

	return true, nil
}

// Has returns true if the key is in the set.
func (m *authenticatedMap[IdentifierType, K, V]) Has(key K) (has bool, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	keyBytes, err := m.keyToBytes(key)
	if err != nil {
		return false, ierrors.Wrap(err, "failed to serialize key")
	}

	return m.has(keyBytes)
}

// Get returns the value for the given key.
func (m *authenticatedMap[IdentifierType, K, V]) Get(key K) (value V, exists bool, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	keyBytes, err := m.keyToBytes(key)
	if err != nil {
		return value, false, ierrors.Wrap(err, "failed to serialize key")
	}

	valueBytes, err := m.tree.Get(keyBytes)
	if err != nil {
		return value, false, ierrors.Wrap(err, "failed to get from tree")
	}

	if valueBytes == nil {
		return value, false, err
	}

	v, consumed, err := m.bytesToValue(valueBytes)
	if err != nil {
		return value, false, ierrors.Wrap(err, "failed to deserialize value")
	}

	if consumed != len(valueBytes) {
		return value, false, ierrors.New("failed to parse entire value")
	}

	return v, true, err
}

// Stream streams all the keys and values.
func (m *authenticatedMap[IdentifierType, K, V]) Stream(callback func(key K, value V) error) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var innerErr error
	if iterationErr := m.rawKeysStore.IterateKeys([]byte{}, func(key K) bool {
		keyBytes, err := m.keyToBytes(key)
		if err != nil {
			innerErr = ierrors.Wrapf(err, "failed to serialize key %s", keyBytes)

			return false
		}

		valueBytes, valueErr := m.tree.Get(keyBytes)
		if valueErr != nil {
			innerErr = ierrors.Wrapf(valueErr, "failed to get value for key %s", keyBytes)

			return false
		}

		value, _, valueErr := m.bytesToValue(valueBytes)
		if valueErr != nil {
			innerErr = ierrors.Wrapf(valueErr, "failed to deserialize value %s", valueBytes)

			return false
		}

		if callbackErr := callback(key, value); callbackErr != nil {
			innerErr = ierrors.Wrapf(callbackErr, "failed to execute callback for key %s", keyBytes)

			return false
		}

		return true
	}); iterationErr != nil {
		return ierrors.Wrap(iterationErr, "failed to iterate over raw keys")
	}

	return innerErr
}

// has returns true if the key is in the map.
func (m *authenticatedMap[IdentifierType, K, V]) has(keyBytes []byte) (has bool, err error) {
	value, err := m.tree.Get(keyBytes)
	if err != nil {
		return false, ierrors.Wrap(err, "failed to get from tree")
	}

	return value != nil, nil
}

func (m *authenticatedMap[IdentifierType, K, V]) addSize(delta int) error {
	size, err := m.size.Get()
	if err != nil && !ierrors.Is(err, kvstore.ErrKeyNotFound) {
		return ierrors.Wrap(err, "failed to get size")
	}

	if err := m.size.Set(uint64(int(size) + delta)); err != nil {
		return ierrors.Wrap(err, "failed to set size")
	}

	return nil
}

package ds

import (
	"crypto/sha256"
	"sync"

	"github.com/pokt-network/smt"

	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/typedkey"
	"github.com/iotaledger/hive.go/lo"
)

const (
	prefixRawKeysStorage uint8 = iota
	prefixRootKey
	prefixSizeKey

	nonEmptyLeaf = 1
)

type authenticatedMap[K, V any] struct {
	rawKeysStore kvstore.KVStore
	tree         *smt.SMT
	size         *typedkey.Number[uint64]
	root         *typedkey.Bytes
	mutex        sync.RWMutex

	kToBytes kvstore.ObjectToBytes[K]
	bytesToK kvstore.BytesToObject[K]
	vToBytes kvstore.ObjectToBytes[V]
	bytesToV kvstore.BytesToObject[V]
}

func newAuthenticatedMap[K, V any](
	store kvstore.KVStore,
	kToBytes kvstore.ObjectToBytes[K],
	bytesToK kvstore.BytesToObject[K],
	vToBytes kvstore.ObjectToBytes[V],
	bytesToV kvstore.BytesToObject[V],
) *authenticatedMap[K, V] {
	newMap := &authenticatedMap[K, V]{
		rawKeysStore: lo.PanicOnErr(store.WithExtendedRealm([]byte{prefixRawKeysStorage})),
		size:         typedkey.NewNumber[uint64](store, prefixSizeKey),
		root:         typedkey.NewBytes(store, prefixRootKey),

		kToBytes: kToBytes,
		bytesToK: bytesToK,
		vToBytes: vToBytes,
		bytesToV: bytesToV,
	}

	if root := newMap.root.Get(); len(root) != 0 {
		newMap.tree = smt.ImportSparseMerkleTree(store, sha256.New(), root, smt.WithValueHasher(nil))
	} else {
		newMap.tree = smt.NewSparseMerkleTree(store, sha256.New(), smt.WithValueHasher(nil))
	}

	return newMap
}

// WasRestoredFromStorage returns true if the map has been restored from storage.
func (m *authenticatedMap[K, V]) WasRestoredFromStorage() bool {
	return len(m.root.Get()) != 0
}

// Root returns the root of the state sparse merkle tree at the latest committed slot.
func (m *authenticatedMap[K, V]) Root() (root types.Identifier) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	copy(root[:], m.tree.Root())

	return
}

// Set sets the output to unspent outputs set.
func (m *authenticatedMap[K, V]) Set(key K, value V) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	valueBytes, err := m.vToBytes(value)
	if err != nil {
		return ierrors.Wrap(err, "failed to serialize value")
	}

	if len(valueBytes) == 0 {
		return ierrors.Errorf("value cannot be empty")
	}

	keyBytes, err := m.kToBytes(key)
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

	if err := m.rawKeysStore.Set(keyBytes, []byte{}); err != nil {
		return ierrors.Wrap(err, "failed to set raw key")
	}

	if !has {
		m.size.Inc()
	}

	return nil
}

// Size returns the number of elements in the map.
func (m *authenticatedMap[K, V]) Size() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return int(m.size.Get())
}

// Commit persists the current state of the map to the storage.
func (m *authenticatedMap[K, V]) Commit() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.root.Set(m.tree.Root())

	return m.tree.Commit()
}

// Delete removes the key from the map.
func (m *authenticatedMap[K, V]) Delete(key K) (deleted bool, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	keyBytes, err := m.kToBytes(key)
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

	if err := m.rawKeysStore.Delete(keyBytes); err != nil {
		return false, ierrors.Wrap(err, "failed to delete from raw keys store")
	}

	if has {
		m.size.Dec()
	}

	return true, nil
}

// Has returns true if the key is in the set.
func (m *authenticatedMap[K, V]) Has(key K) (has bool, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	keyBytes, err := m.kToBytes(key)
	if err != nil {
		return false, ierrors.Wrap(err, "failed to serialize key")
	}

	return m.has(keyBytes)
}

// Get returns the value for the given key.
func (m *authenticatedMap[K, V]) Get(key K) (value V, exists bool, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	keyBytes, err := m.kToBytes(key)
	if err != nil {
		return value, false, ierrors.Wrap(err, "failed to serialize key")
	}

	valueBytes, err := m.tree.Get(keyBytes)
	if err != nil {
		if ierrors.Is(err, kvstore.ErrKeyNotFound) {
			return value, false, err
		}

		return value, false, ierrors.Wrap(err, "failed to get from tree")
	}

	if len(valueBytes) == 0 {
		return value, false, err
	}

	v, consumed, err := m.bytesToV(valueBytes)
	if err != nil {
		return value, false, ierrors.Wrap(err, "failed to deserialize value")
	}

	if consumed != len(valueBytes) {
		return value, false, ierrors.New("failed to parse entire value")
	}

	return v, true, err
}

// Stream streams all the keys and values.
func (m *authenticatedMap[K, V]) Stream(callback func(key K, value V) error) (err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if iterationErr := m.rawKeysStore.Iterate([]byte{}, func(key kvstore.Key, _ kvstore.Value) bool {
		valueBytes, valueErr := m.tree.Get(key)
		if valueErr != nil {
			err = ierrors.Wrapf(valueErr, "failed to get value for key %s", key)

			return false
		}

		k, _, keyErr := m.bytesToK(key)
		if keyErr != nil {
			err = ierrors.Wrapf(keyErr, "failed to deserialize key %s", key)

			return false
		}

		v, _, valueErr := m.bytesToV(valueBytes)
		if valueErr != nil {
			err = ierrors.Wrapf(valueErr, "failed to deserialize value %s", valueBytes)

			return false
		}

		if callbackErr := callback(k, v); callbackErr != nil {
			err = ierrors.Wrapf(callbackErr, "failed to execute callback for key %s", key)

			return false
		}

		return true
	}); iterationErr != nil {
		err = ierrors.Wrap(iterationErr, "failed to iterate over raw keys")
	}

	return
}

// has returns true if the key is in the map.
func (m *authenticatedMap[K, V]) has(keyBytes []byte) (has bool, err error) {
	value, err := m.tree.Get(keyBytes)
	if err != nil {
		return false, ierrors.Wrap(err, "failed to get from tree")
	}

	return value != nil, nil
}

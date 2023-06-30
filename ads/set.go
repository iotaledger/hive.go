package ads

import (
	"sync"

	"github.com/celestiaorg/smt"
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/typedkey"
	"github.com/iotaledger/hive.go/lo"
)

const (
	PrefixRawKeysStorage uint8 = iota
	PrefixSMTKeysStorage
	PrefixSMTValuesStorage
	PrefixRootKey
	PrefixSizeKey

	nonEmptyLeaf = 1
)

// Set is a sparse merkle tree based set.
type Set[K any] struct {
	rawKeysStore kvstore.KVStore
	tree         *smt.SparseMerkleTree
	root         *typedkey.Bytes
	size         *typedkey.Number[uint64]
	mutex        sync.Mutex

	kToBytes ObjectToBytes[K]
	bytesToK BytesToObject[K]
}

// NewSet creates a new sparse merkle tree based set.
func NewSet[K any](
	store kvstore.KVStore,
	kToBytes ObjectToBytes[K],
	bytesToK BytesToObject[K],
) (newSet *Set[K]) {
	newSet = &Set[K]{
		rawKeysStore: lo.PanicOnErr(store.WithExtendedRealm([]byte{PrefixRawKeysStorage})),
		tree: smt.NewSparseMerkleTree(
			lo.PanicOnErr(store.WithExtendedRealm([]byte{PrefixSMTKeysStorage})),
			lo.PanicOnErr(store.WithExtendedRealm([]byte{PrefixSMTValuesStorage})),
			lo.PanicOnErr(blake2b.New256(nil)),
		),
		root:     typedkey.NewBytes(store, PrefixRootKey),
		size:     typedkey.NewNumber[uint64](store, PrefixSizeKey),
		kToBytes: kToBytes,
		bytesToK: bytesToK,
	}

	if root := newSet.root.Get(); len(root) != 0 {
		newSet.tree.SetRoot(root)
	}

	return
}

func (s *Set[K]) IsNew() bool {
	return len(s.root.Get()) == 0
}

// Root returns the root of the state sparse merkle tree at the latest committed slot.
func (s *Set[K]) Root() (root types.Identifier) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	copy(root[:], s.tree.Root())

	return
}

// Add adds the key to the set.
func (s *Set[K]) Add(key K) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	keyBytes := lo.PanicOnErr(s.kToBytes(key))
	if s.has(keyBytes) {
		return
	}

	s.root.Set(lo.PanicOnErr(s.tree.Update(keyBytes, []byte{nonEmptyLeaf})))
	s.size.Inc()

	if err := s.rawKeysStore.Set(keyBytes, []byte{}); err != nil {
		panic(err)
	}
}

// Delete removes the key from the set.
func (s *Set[K]) Delete(key K) (deleted bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	keyBytes := lo.PanicOnErr(s.kToBytes(key))
	if deleted = s.has(keyBytes); !deleted {
		return
	}

	s.root.Set(lo.PanicOnErr(s.tree.Delete(keyBytes)))
	s.size.Dec()

	if err := s.rawKeysStore.Delete(keyBytes); err != nil {
		panic(err)
	}

	return
}

// Has returns true if the key is in the set.
func (s *Set[K]) Has(key K) (has bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.has(lo.PanicOnErr(s.kToBytes(key)))
}

// Stream iterates over the set and calls the callback for each element.
func (s *Set[K]) Stream(callback func(key K) bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var innerErr error
	if iterationErr := s.rawKeysStore.Iterate([]byte{}, func(key kvstore.Key, _ kvstore.Value) bool {
		k, _, keyErr := s.bytesToK(key)
		if keyErr != nil {
			innerErr = ierrors.Wrapf(keyErr, "failed to deserialize key %s", key)
			return false
		}

		return callback(k)
	}); iterationErr != nil {
		return ierrors.Wrap(iterationErr, "failed to iterate over set members")
	}

	return innerErr
}

// Size returns the number of elements in the set.
func (s *Set[K]) Size() (size int) {
	return int(s.size.Get())
}

// has returns true if the key is in the set.
func (s *Set[K]) has(key []byte) (has bool) {
	has, err := s.tree.Has(key)
	if err != nil {
		if ierrors.Is(err, kvstore.ErrKeyNotFound) {
			return false
		}

		panic(err)
	}

	return
}

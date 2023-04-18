package account

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/ads"
	"github.com/iotaledger/hive.go/core/storable"
	"github.com/iotaledger/hive.go/crypto/identity"
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/kvstore"
)

const cacheSize = 1000

// Accounts is a mapping between a collection of identities and their weights.
type Accounts struct {
	weights     *ads.Map[identity.ID, storable.SerializableInt64, *identity.ID, *storable.SerializableInt64]
	cacheMutex  sync.Mutex
	totalWeight int64
	mutex       sync.RWMutex
}

// NewAccounts creates a new Weights instance.
func NewAccounts(store kvstore.KVStore) *Accounts {
	newWeights := &Accounts{
		weights:     ads.NewMap[identity.ID, storable.SerializableInt64](store),
		totalWeight: 0,
	}

	if err := newWeights.weights.Stream(func(_ identity.ID, value *storable.SerializableInt64) bool {
		newWeights.totalWeight += int64(*value)
		return true
	}); err != nil {
		return nil
	}

	return newWeights
}

// SelectAccounts creates a new WeightedSet instance, that maintains a correct and updated total weight of its members.
func (w *Accounts) SelectAccounts(members ...identity.ID) (selectedAccounts *SelectedAccounts) {
	return NewSelectedAccounts(w, members...)
}

// Get returns the weight of the given identity.
func (w *Accounts) Get(id identity.ID) (weight int64, exists bool) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	serializedWeight, exists := w.weights.Get(id)
	if !exists {
		return 0, false
	}

	return int64(*serializedWeight), exists
}

// Update updates the weight of the given identity.
func (w *Accounts) Update(id identity.ID, weightDiff int64) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	newWeight := storable.SerializableInt64(weightDiff)
	oldWeight, _ := w.weights.Get(id)
	w.weights.Set(id, newWeight.Add(oldWeight))

	w.totalWeight += weightDiff
}

// ForEach iterates over all weights and calls the given callback for each of them.
func (w *Accounts) ForEach(callback func(id identity.ID, weight int64) bool) (err error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.weights.Stream(func(id identity.ID, weight *storable.SerializableInt64) bool { return callback(id, int64(*weight)) })
}

// TotalWeight returns the total weight of all identities.
func (w *Accounts) TotalWeight() (totalWeight int64) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.totalWeight
}

// Root returns the root of the merkle tree of the stored weights.
func (w *Accounts) Root() (root types.Identifier) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.weights.Root()
}

// Map returns the weights as a map.
func (w *Accounts) Map() (weights map[identity.ID]int64, err error) {
	weights = make(map[identity.ID]int64)
	if err = w.ForEach(func(id identity.ID, weight int64) bool {
		weights[id] = weight
		return true
	}); err != nil {
		return nil, errors.Wrap(err, "failed to export weights")
	}

	return weights, nil
}

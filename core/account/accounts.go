package account

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/ads"
	"github.com/iotaledger/hive.go/core/storable"
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/serializer/v2"
)

type AccountIDType interface {
	comparable
	serializer.Byter
}

// Accounts is a mapping between a collection of identities and their weights.
type Accounts[AccountID AccountIDType, AccountIDPtr serializer.MarshalablePtr[AccountID]] struct {
	weights     *ads.Map[AccountID, storable.SerializableInt64, AccountIDPtr, *storable.SerializableInt64]
	totalWeight int64
	mutex       sync.RWMutex
}

// NewAccounts creates a new Weights instance.
func NewAccounts[A AccountIDType, APtr serializer.MarshalablePtr[A]](store kvstore.KVStore) *Accounts[A, APtr] {
	newAccounts := &Accounts[A, APtr]{
		weights:     ads.NewMap[A, storable.SerializableInt64, APtr](store),
		totalWeight: 0,
	}

	if err := newAccounts.weights.Stream(func(_ A, value *storable.SerializableInt64) bool {
		newAccounts.totalWeight += int64(*value)
		return true
	}); err != nil {
		return nil
	}

	return newAccounts
}

// SelectAccounts creates a new WeightedSet instance, that maintains a correct and updated total weight of its members.
func (w *Accounts[AccountID, AccountIDPtr]) SelectAccounts(members ...AccountID) (selectedAccounts *SelectedAccounts[AccountID, AccountIDPtr]) {
	return NewSelectedAccounts(w, members...)
}

// Get returns the weight of the given identity.
func (w *Accounts[AccountID, AccountIDPtr]) Get(id AccountID) (weight int64, exists bool) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	serializedWeight, exists := w.weights.Get(id)
	if !exists {
		return 0, false
	}

	return int64(*serializedWeight), exists
}

// Set sets the weight of the given identity.
func (w *Accounts[AccountID, AccountIDPtr]) Set(id AccountID, weight int64) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	oldWeight, exists := w.weights.Get(id)
	if exists {
		w.totalWeight -= int64(*oldWeight)
	}

	newWeight := storable.SerializableInt64(weight)
	w.weights.Set(id, &newWeight)
	w.totalWeight += weight
}

// ForEach iterates over all weights and calls the given callback for each of them.
func (w *Accounts[AccountID, AccountIDPtr]) ForEach(callback func(id AccountID, weight int64) bool) (err error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.weights.Stream(func(id AccountID, weight *storable.SerializableInt64) bool { return callback(id, int64(*weight)) })
}

// TotalWeight returns the total weight of all identities.
func (w *Accounts[AccountID, AccountIDPtr]) TotalWeight() (totalWeight int64) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.totalWeight
}

// Root returns the root of the merkle tree of the stored weights.
func (w *Accounts[AccountID, AccountIDPtr]) Root() (root types.Identifier) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.weights.Root()
}

// Map returns the weights as a map.
func (w *Accounts[AccountID, AccountIDPtr]) Map() (weights map[AccountID]int64, err error) {
	weights = make(map[AccountID]int64)
	if err = w.ForEach(func(id AccountID, weight int64) bool {
		weights[id] = weight
		return true
	}); err != nil {
		return nil, errors.Wrap(err, "failed to export weights")
	}

	return weights, nil
}

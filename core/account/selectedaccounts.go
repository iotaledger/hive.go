package account

import (
	"sync"

	"github.com/iotaledger/hive.go/core/index"
	"github.com/iotaledger/hive.go/crypto/identity"
	"github.com/iotaledger/hive.go/ds/advancedset"
	"github.com/iotaledger/hive.go/runtime/event"
)

type SelectedAccounts[I index.Type] struct {
	Weights             *Accounts[I]
	weightUpdatesDetach *event.Hook[func(*AccountsUpdateBatch[I])]
	members             *advancedset.AdvancedSet[identity.ID]
	membersMutex        sync.RWMutex
	totalWeight         int64
	totalWeightMutex    sync.RWMutex
}

func NewSelectedAccounts[I index.Type](accounts *Accounts[I], optMembers ...identity.ID) *SelectedAccounts[I] {
	newWeightedSet := new(SelectedAccounts[I])
	newWeightedSet.Weights = accounts
	newWeightedSet.members = advancedset.New[identity.ID]()

	newWeightedSet.weightUpdatesDetach = accounts.events.WeightsUpdated.Hook(newWeightedSet.onWeightUpdated)

	for _, member := range optMembers {
		newWeightedSet.Add(member)
	}

	return newWeightedSet
}

func (w *SelectedAccounts[I]) Add(id identity.ID) (added bool) {
	w.Weights.mutex.RLock()
	defer w.Weights.mutex.RUnlock()

	w.membersMutex.Lock()
	defer w.membersMutex.Unlock()

	w.totalWeightMutex.Lock()
	defer w.totalWeightMutex.Unlock()

	if added = w.members.Add(id); added {
		if weight, exists := w.Weights.get(id); exists {
			w.totalWeight += weight.Value
		}
	}

	return
}

func (w *SelectedAccounts[I]) Delete(id identity.ID) (removed bool) {
	w.Weights.mutex.RLock()
	defer w.Weights.mutex.RUnlock()

	w.membersMutex.Lock()
	defer w.membersMutex.Unlock()

	w.totalWeightMutex.Lock()
	defer w.totalWeightMutex.Unlock()

	if removed = w.members.Delete(id); removed {
		if weight, exists := w.Weights.get(id); exists {
			w.totalWeight -= weight.Value
		}
	}

	return
}

func (w *SelectedAccounts[I]) Get(id identity.ID) (weight *Weight[I], exists bool) {
	w.membersMutex.RLock()
	defer w.membersMutex.RUnlock()

	if !w.members.Has(id) {
		return nil, false
	}

	if weight, exists = w.Weights.Get(id); exists {
		return weight, true
	}

	return NewWeight[I](0, -1), true
}

func (w *SelectedAccounts[I]) Has(id identity.ID) (has bool) {
	w.membersMutex.RLock()
	defer w.membersMutex.RUnlock()

	return w.members.Has(id)
}

// TODO: do we actually need two foreaches?
func (w *SelectedAccounts[I]) ForEach(callback func(id identity.ID) error) (err error) {
	for it := w.members.Iterator(); it.HasNext(); {
		member := it.Next()
		if err = callback(member); err != nil {
			return
		}
	}

	return
}

func (w *SelectedAccounts[I]) ForEachWeighted(callback func(id identity.ID, weight int64) error) (err error) {
	for it := w.members.Iterator(); it.HasNext(); {
		member := it.Next()
		memberWeight, exists := w.Weights.Get(member)
		if !exists {
			memberWeight = NewWeight[I](0, -1)
		}
		if err = callback(member, memberWeight.Value); err != nil {
			return
		}
	}

	return
}

func (w *SelectedAccounts[I]) TotalWeight() (totalWeight int64) {
	w.totalWeightMutex.RLock()
	defer w.totalWeightMutex.RUnlock()

	return w.totalWeight
}

func (w *SelectedAccounts[I]) Identities() *advancedset.AdvancedSet[identity.ID] {
	w.membersMutex.RLock()
	defer w.membersMutex.RUnlock()

	return w.members
}

func (w *SelectedAccounts[I]) Detach() {
	w.weightUpdatesDetach.Unhook()
}

func (w *SelectedAccounts[I]) onWeightUpdated(updates *AccountsUpdateBatch[I]) {
	w.totalWeightMutex.Lock()
	defer w.totalWeightMutex.Unlock()

	updates.ForEach(func(id identity.ID, diff int64) {
		if w.members.Has(id) {
			w.totalWeight += diff
		}
	})
}

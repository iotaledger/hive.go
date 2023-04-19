package account

import (
	"sync"

	"github.com/iotaledger/hive.go/crypto/identity"
	"github.com/iotaledger/hive.go/ds/advancedset"
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/lo"
)

type SelectedAccounts struct {
	accounts         *Accounts
	members          *shrinkingmap.ShrinkingMap[identity.ID, types.Empty]
	totalWeight      int64
	totalWeightMutex sync.RWMutex
}

func NewSelectedAccounts(accounts *Accounts, optMembers ...identity.ID) *SelectedAccounts {
	newWeightedSet := new(SelectedAccounts)
	newWeightedSet.accounts = accounts
	newWeightedSet.members = shrinkingmap.New[identity.ID, types.Empty]()

	for _, member := range optMembers {
		newWeightedSet.Add(member)
	}

	return newWeightedSet
}

func (w *SelectedAccounts) Add(id identity.ID) (added bool) {
	w.accounts.mutex.RLock()
	defer w.accounts.mutex.RUnlock()

	w.totalWeightMutex.Lock()
	defer w.totalWeightMutex.Unlock()

	if added = w.members.Set(id, types.Void); added {
		if weight, exists := w.accounts.Get(id); exists {
			w.totalWeight += weight
		}
	}

	return
}

func (w *SelectedAccounts) Delete(id identity.ID) (removed bool) {
	w.accounts.mutex.RLock()
	defer w.accounts.mutex.RUnlock()

	w.totalWeightMutex.Lock()
	defer w.totalWeightMutex.Unlock()

	if removed = w.members.Delete(id); removed {
		if weight, exists := w.accounts.Get(id); exists {
			w.totalWeight -= weight
		}
	}

	return
}

func (w *SelectedAccounts) Get(id identity.ID) (weight int64, exists bool) {
	// check if the member is part of the committee, otherwise its weight is 0
	if !w.members.Has(id) {
		return 0, false
	}

	if weight, exists = w.accounts.Get(id); exists {
		return weight, true
	}

	return 0, true
}

func (w *SelectedAccounts) Has(id identity.ID) (has bool) {
	return w.members.Has(id)
}

func (w *SelectedAccounts) ForEach(callback func(id identity.ID, weight int64) error) (err error) {
	w.members.ForEachKey(func(member identity.ID) bool {
		if err := callback(member, lo.Return1(w.accounts.Get(member))); err != nil {
			return false
		}

		return true
	})

	return
}

func (w *SelectedAccounts) TotalWeight() (totalWeight int64) {
	w.totalWeightMutex.RLock()
	defer w.totalWeightMutex.RUnlock()

	return w.totalWeight
}

func (w *SelectedAccounts) Members() *advancedset.AdvancedSet[identity.ID] {
	return advancedset.New(w.members.Keys()...)
}

func (w *SelectedAccounts) SelectAccounts(members ...identity.ID) *SelectedAccounts {
	var selectedMembers []identity.ID
	for _, member := range members {
		if w.members.Has(member) {
			selectedMembers = append(selectedMembers, member)
		}
	}

	return NewSelectedAccounts(w.accounts, selectedMembers...)
}

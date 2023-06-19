package account

import (
	"sync"

	"github.com/iotaledger/hive.go/ds/advancedset"
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
)

type SelectedAccounts[AccountID AccountIDType, AccountIDPtr serializer.MarshalablePtr[AccountID]] struct {
	accounts         *Accounts[AccountID, AccountIDPtr]
	members          *shrinkingmap.ShrinkingMap[AccountID, types.Empty]
	totalWeight      int64
	totalWeightMutex sync.RWMutex
}

func NewSelectedAccounts[A AccountIDType, APtr serializer.MarshalablePtr[A]](accounts *Accounts[A, APtr], optMembers ...A) *SelectedAccounts[A, APtr] {
	newWeightedSet := new(SelectedAccounts[A, APtr])
	newWeightedSet.accounts = accounts
	newWeightedSet.members = shrinkingmap.New[A, types.Empty]()

	for _, member := range optMembers {
		newWeightedSet.Add(member)
	}

	return newWeightedSet
}

func (w *SelectedAccounts[AccountID, AccountIDPtr]) Add(id AccountID) (added bool) {
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

func (w *SelectedAccounts[AccountID, AccountIDPtr]) Delete(id AccountID) (removed bool) {
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

func (w *SelectedAccounts[AccountID, AccountIDPtr]) Get(id AccountID) (weight int64, exists bool) {
	// check if the member is part of the committee, otherwise its weight is 0
	if !w.members.Has(id) {
		return 0, false
	}

	if weight, exists = w.accounts.Get(id); exists {
		return weight, true
	}

	return 0, true
}

func (w *SelectedAccounts[AccountID, AccountIDPtr]) Has(id AccountID) (has bool) {
	return w.members.Has(id)
}

func (w *SelectedAccounts[AccountID, AccountIDPtr]) ForEach(callback func(id AccountID, weight int64) error) (err error) {
	w.members.ForEachKey(func(member AccountID) bool {
		if err = callback(member, lo.Return1(w.accounts.Get(member))); err != nil {
			return false
		}

		return true
	})

	return
}

func (w *SelectedAccounts[AccountID, AccountIDPtr]) TotalWeight() (totalWeight int64) {
	w.totalWeightMutex.RLock()
	defer w.totalWeightMutex.RUnlock()

	return w.totalWeight
}

func (w *SelectedAccounts[AccountID, AccountIDPtr]) Members() *advancedset.AdvancedSet[AccountID] {
	return advancedset.New(w.members.Keys()...)
}

func (w *SelectedAccounts[AccountID, AccountIDPtr]) SelectAccounts(members ...AccountID) *SelectedAccounts[AccountID, AccountIDPtr] {
	var selectedMembers []AccountID
	for _, member := range members {
		if w.members.Has(member) {
			selectedMembers = append(selectedMembers, member)
		}
	}

	return NewSelectedAccounts(w.accounts, selectedMembers...)
}

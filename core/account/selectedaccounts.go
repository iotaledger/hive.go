package account

import (
	"fmt"

	"github.com/iotaledger/hive.go/ds/advancedset"
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/serializer/v2"
)

type SeatIndex int

type SelectedAccounts[AccountID AccountIDType, AccountIDPtr serializer.MarshalablePtr[AccountID]] struct {
	accounts       *Accounts[AccountID, AccountIDPtr]
	seatsByAccount *shrinkingmap.ShrinkingMap[AccountID, SeatIndex]
	accountsBySeat *shrinkingmap.ShrinkingMap[SeatIndex, AccountID]
}

func NewSelectedAccounts[A AccountIDType, APtr serializer.MarshalablePtr[A]](accounts *Accounts[A, APtr], optMembers ...A) *SelectedAccounts[A, APtr] {
	newWeightedSet := new(SelectedAccounts[A, APtr])
	newWeightedSet.accounts = accounts
	newWeightedSet.seatsByAccount = shrinkingmap.New[A, SeatIndex]()
	newWeightedSet.accountsBySeat = shrinkingmap.New[SeatIndex, A]()

	for i, member := range optMembers {
		newWeightedSet.seatsByAccount.Set(member, SeatIndex(i))
		newWeightedSet.accountsBySeat.Set(SeatIndex(i), member)
	}

	return newWeightedSet
}

func (w *SelectedAccounts[AccountID, AccountIDPtr]) Set(seat SeatIndex, id AccountID) {
	if oldSeat, exists := w.seatsByAccount.Get(id); exists {
		if oldSeat != seat {
			panic(fmt.Sprintf("account already selected with a different seat: %d vs %d", oldSeat, seat))
		}
	}

	w.seatsByAccount.Set(id, seat)
	w.accountsBySeat.Set(seat, id)
}

func (w *SelectedAccounts[AccountID, AccountIDPtr]) Delete(id AccountID) {
	if oldSeat, exists := w.seatsByAccount.Get(id); exists {
		w.seatsByAccount.Delete(id)
		w.accountsBySeat.Delete(oldSeat)
	}
}

func (w *SelectedAccounts[AccountID, AccountIDPtr]) GetSeat(id AccountID) (seat SeatIndex, exists bool) {
	return w.seatsByAccount.Get(id)
}

func (w *SelectedAccounts[AccountID, AccountIDPtr]) HasAccount(id AccountID) (has bool) {
	return w.seatsByAccount.Has(id)
}

func (w *SelectedAccounts[AccountID, AccountIDPtr]) SeatCount() int {
	return w.accountsBySeat.Size()
}

func (w *SelectedAccounts[AccountID, AccountIDPtr]) Accounts() *advancedset.AdvancedSet[AccountID] {
	return advancedset.New(w.seatsByAccount.Keys()...)
}

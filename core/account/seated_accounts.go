package account

import (
	"github.com/iotaledger/hive.go/ds/advancedset"
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/serializer/v2"
)

type SeatIndex int

type SeatedAccounts[AccountID AccountIDType, AccountIDPtr serializer.MarshalablePtr[AccountID]] struct {
	accounts       *Accounts[AccountID, AccountIDPtr]
	seatsByAccount *shrinkingmap.ShrinkingMap[AccountID, SeatIndex]
	accountsBySeat *shrinkingmap.ShrinkingMap[SeatIndex, AccountID]
}

func NewSeatedAccounts[A AccountIDType, APtr serializer.MarshalablePtr[A]](accounts *Accounts[A, APtr], optMembers ...A) *SeatedAccounts[A, APtr] {
	newWeightedSet := new(SeatedAccounts[A, APtr])
	newWeightedSet.accounts = accounts
	newWeightedSet.seatsByAccount = shrinkingmap.New[A, SeatIndex]()
	newWeightedSet.accountsBySeat = shrinkingmap.New[SeatIndex, A]()

	for i, member := range optMembers {
		newWeightedSet.seatsByAccount.Set(member, SeatIndex(i))
		newWeightedSet.accountsBySeat.Set(SeatIndex(i), member)
	}

	return newWeightedSet
}

func (w *SeatedAccounts[AccountID, AccountIDPtr]) Set(seat SeatIndex, id AccountID) bool {
	if _, exists := w.accounts.Get(id); !exists {
		return false
	}

	if oldSeat, exists := w.seatsByAccount.Get(id); exists {
		if oldSeat != seat {
			return false
		}
	}

	w.seatsByAccount.Set(id, seat)
	w.accountsBySeat.Set(seat, id)

	return true
}

func (w *SeatedAccounts[AccountID, AccountIDPtr]) Delete(id AccountID) bool {
	if oldSeat, exists := w.seatsByAccount.Get(id); exists {
		w.seatsByAccount.Delete(id)
		w.accountsBySeat.Delete(oldSeat)

		return true
	}

	return false
}

func (w *SeatedAccounts[AccountID, AccountIDPtr]) GetSeat(id AccountID) (seat SeatIndex, exists bool) {
	return w.seatsByAccount.Get(id)
}

func (w *SeatedAccounts[AccountID, AccountIDPtr]) HasAccount(id AccountID) (has bool) {
	return w.seatsByAccount.Has(id)
}

func (w *SeatedAccounts[AccountID, AccountIDPtr]) SeatCount() int {
	return w.accountsBySeat.Size()
}

func (w *SeatedAccounts[AccountID, AccountIDPtr]) Accounts() *advancedset.AdvancedSet[AccountID] {
	return advancedset.New(w.seatsByAccount.Keys()...)
}

package account

import (
	"github.com/iotaledger/hive.go/ds/advancedset"
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/serializer/v2"
)

type SeatIndex int

type SeatedAccounts[AccountID IDType, AccountIDPtr serializer.MarshalablePtr[AccountID]] struct {
	accounts       *Accounts[AccountID, AccountIDPtr]
	seatsByAccount *shrinkingmap.ShrinkingMap[AccountID, SeatIndex]
}

func NewSeatedAccounts[A IDType, APtr serializer.MarshalablePtr[A]](accounts *Accounts[A, APtr], optMembers ...A) *SeatedAccounts[A, APtr] {
	s := &SeatedAccounts[A, APtr]{
		accounts:       accounts,
		seatsByAccount: shrinkingmap.New[A, SeatIndex](),
	}

	for i, member := range optMembers {
		s.seatsByAccount.Set(member, SeatIndex(i))
	}

	return s
}

func (s *SeatedAccounts[AccountID, AccountIDPtr]) Set(seat SeatIndex, id AccountID) bool {
	// Check if the account exists.
	if _, exists := s.accounts.Get(id); !exists {
		return false
	}

	// Check if the account already has a seat.
	if oldSeat, exists := s.seatsByAccount.Get(id); exists {
		if oldSeat != seat {
			return false
		}
	}

	return s.seatsByAccount.Set(id, seat)
}

func (s *SeatedAccounts[AccountID, AccountIDPtr]) Delete(id AccountID) bool {
	return s.seatsByAccount.Delete(id)
}

func (s *SeatedAccounts[AccountID, AccountIDPtr]) GetSeat(id AccountID) (seat SeatIndex, exists bool) {
	return s.seatsByAccount.Get(id)
}

func (s *SeatedAccounts[AccountID, AccountIDPtr]) HasAccount(id AccountID) (has bool) {
	return s.seatsByAccount.Has(id)
}

func (s *SeatedAccounts[AccountID, AccountIDPtr]) SeatCount() int {
	return s.seatsByAccount.Size()
}

func (s *SeatedAccounts[AccountID, AccountIDPtr]) Accounts() *advancedset.AdvancedSet[AccountID] {
	return advancedset.New(s.seatsByAccount.Keys()...)
}

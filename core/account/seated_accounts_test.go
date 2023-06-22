package account_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/account"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
)

func TestSelectedAccounts(t *testing.T) {
	store := mapdb.NewMapDB()

	// Create a new set of accounts
	accounts := account.NewAccounts[testID](store)

	// Add some accounts
	account1 := testID([]byte{1})
	account2 := testID([]byte{2})
	account3 := testID([]byte{3})
	account4 := testID([]byte{4})

	accounts.Set(account1, 10)
	accounts.Set(account2, 20)
	accounts.Set(account3, 30)

	// Create a new set of selected accounts
	seatedAccounts := account.NewSeatedAccounts(accounts, account1, account3)
	require.Equal(t, 2, seatedAccounts.SeatCount())

	// Test the "Set" method
	added := seatedAccounts.Set(account.SeatIndex(3), account2)
	require.True(t, added)
	require.Equal(t, 3, seatedAccounts.SeatCount())

	// Try adding an account again with a different seat.
	added = seatedAccounts.Set(account.SeatIndex(2), account2)
	require.False(t, added)
	require.Equal(t, 3, seatedAccounts.SeatCount())

	// Try adding an account that does not exist in accounts.
	added = seatedAccounts.Set(account.SeatIndex(4), account4)
	require.False(t, added)
	require.Equal(t, 3, seatedAccounts.SeatCount())

	// Test the "Delete" method
	removed := seatedAccounts.Delete(account1)
	require.True(t, removed)
	require.Equal(t, 2, seatedAccounts.SeatCount())

	// Test the "Get" method
	seat, exists := seatedAccounts.GetSeat(account2)
	require.True(t, exists)
	require.Equal(t, account.SeatIndex(3), seat)

	// Test the "Get" method with account that's not in accounts.
	_, exists = seatedAccounts.GetSeat(account4)
	require.False(t, exists)

	// Test the "Has" method
	has := seatedAccounts.HasAccount(account3)
	require.True(t, has)

	// Test the "Members" method
	members := seatedAccounts.Accounts()
	require.Equal(t, 2, members.Size())
	require.True(t, members.Has(account2))
	require.True(t, members.Has(account3))
}

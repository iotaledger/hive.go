package account_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/hive.go/core/account"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/stretchr/testify/require"
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
	selectedAccounts := account.NewSelectedAccounts(accounts, account1, account3)

	// Test the "Add" method
	added := selectedAccounts.Add(account2)
	require.True(t, added)
	require.Equal(t, int64(60), selectedAccounts.TotalWeight())

	// Add an account that does not exist in accouts. Total weight should be the same
	added = selectedAccounts.Add(account4)
	require.True(t, added)
	require.Equal(t, int64(60), selectedAccounts.TotalWeight())

	// Test the "Delete" method
	removed := selectedAccounts.Delete(account1)
	require.True(t, removed)
	require.Equal(t, int64(50), selectedAccounts.TotalWeight())

	// Test the "Get" method
	weight, exists := selectedAccounts.Get(account2)
	require.True(t, exists)
	require.Equal(t, int64(20), weight)

	// Test the "Get" method with account that's not in accounts.
	weight, exists = selectedAccounts.Get(account4)
	require.True(t, exists)
	require.Equal(t, int64(0), weight)

	// Test Get non-existed account
	weight, exists = selectedAccounts.Get(account1)
	require.False(t, exists)
	require.Equal(t, int64(0), weight)

	// Test the "Has" method
	has := selectedAccounts.Has(account3)
	require.True(t, has)

	// Test the "ForEach" method
	totalWeight := int64(0)
	err := selectedAccounts.ForEach(func(id testID, weight int64) error {
		totalWeight += weight
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, int64(50), totalWeight)

	// Test the "ForEach" method, with error in callback function
	err = selectedAccounts.ForEach(func(id testID, weight int64) error {
		return errors.New("error!!")
	})
	require.Error(t, err)
	require.EqualError(t, err, "error!!")

	// Test the "Members" method
	members := selectedAccounts.Members()
	require.Equal(t, 3, members.Size())
	require.True(t, members.Has(account2))
	require.True(t, members.Has(account3))
	require.True(t, members.Has(account4))

	// Test the "SelectAccounts" method
	selectedAccounts2 := selectedAccounts.SelectAccounts(account3)
	require.Equal(t, int64(30), selectedAccounts2.TotalWeight())
	require.True(t, selectedAccounts2.Has(account3))
	require.False(t, selectedAccounts2.Has(account1))
}

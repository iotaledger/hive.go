package account_test

import (
	"testing"

	"github.com/iotaledger/hive.go/core/account"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/stretchr/testify/assert"
)

func TestSelectedAccounts(t *testing.T) {
	store := mapdb.NewMapDB()

	// Create a new set of accounts
	accounts := account.NewAccounts[testID](store)

	// Add some accounts
	account1 := testID([]byte{1})
	account2 := testID([]byte{2})
	account3 := testID([]byte{3})

	accounts.Set(account1, 10)
	accounts.Set(account2, 20)
	accounts.Set(account3, 30)

	// Create a new set of selected accounts
	selectedAccounts := account.NewSelectedAccounts(accounts, account1, account3)

	// Test the "Add" method
	added := selectedAccounts.Add(account2)
	assert.True(t, added)
	assert.Equal(t, int64(60), selectedAccounts.TotalWeight())

	// Test the "Delete" method
	removed := selectedAccounts.Delete(account1)
	assert.True(t, removed)
	assert.Equal(t, int64(50), selectedAccounts.TotalWeight())

	// Test the "Get" method
	weight, exists := selectedAccounts.Get(account2)
	assert.True(t, exists)
	assert.Equal(t, int64(20), weight)

	// Test the "Has" method
	has := selectedAccounts.Has(account3)
	assert.True(t, has)

	// Test the "ForEach" method
	totalWeight := int64(0)
	selectedAccounts.ForEach(func(id testID, weight int64) error {
		totalWeight += weight
		return nil
	})
	assert.Equal(t, int64(50), totalWeight)

	// Test the "Members" method
	members := selectedAccounts.Members()
	assert.Equal(t, 2, members.Size())
	assert.True(t, members.Has(account2))
	assert.True(t, members.Has(account3))

	// Test the "SelectAccounts" method
	selectedAccounts2 := selectedAccounts.SelectAccounts(account3)
	assert.Equal(t, int64(30), selectedAccounts2.TotalWeight())
	assert.True(t, selectedAccounts2.Has(account3))
	assert.False(t, selectedAccounts2.Has(account1))
}

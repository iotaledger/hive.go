package account_test

import (
	"testing"

	"github.com/iotaledger/hive.go/core/account"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/stretchr/testify/require"
)

func TestAccounts(t *testing.T) {
	store := mapdb.NewMapDB()
	accounts := account.NewAccounts[testID](store)
	issuers, totalWeight := generateAccounts()

	// Add accounts
	for id, weight := range issuers {
		accounts.Set(id, weight)
	}

	// Test Map
	wMap, err := accounts.Map()
	require.NoError(t, err)
	require.EqualValues(t, issuers, wMap)
	require.Equal(t, totalWeight, accounts.TotalWeight())

	// update issuer's weight
	issuerIDs := lo.Keys(wMap)
	oldWeight, exist := accounts.Get(issuerIDs[0])
	require.True(t, exist)

	newWeight := int64(20)
	accounts.Set(issuerIDs[0], newWeight)
	w, exist := accounts.Get(issuerIDs[0])
	require.True(t, exist)
	require.Equal(t, newWeight, w)
	require.Equal(t, totalWeight-oldWeight+newWeight, accounts.TotalWeight())

	// Get a non existed account
	_, exist = accounts.Get(testID([]byte{tpkg.RandByte()}))
	require.False(t, exist)

	// Test Selected Accounts, get 1 issuer
	selected := accounts.SelectAccounts(issuerIDs[0])
	require.Equal(t, newWeight, selected.TotalWeight())
	require.Equal(t, 1, selected.Members().Size())

	// Test Selected Accounts, get all issuers
	selected = accounts.SelectAccounts(issuerIDs...)
	require.Equal(t, accounts.TotalWeight(), selected.TotalWeight())
	require.Equal(t, len(issuerIDs), selected.Members().Size())

	root := accounts.Root()

	// load from existing store
	accounts1 := account.NewAccounts[testID](store)
	require.Equal(t, root, accounts1.Root())
	require.Equal(t, accounts.TotalWeight(), accounts1.TotalWeight())
}

type testID [1]byte

func (t testID) Bytes() ([]byte, error) {
	return t[:], nil
}

func (t *testID) FromBytes(b []byte) (int, error) {
	copy(t[:], b)
	return len(t), nil
}

func generateAccounts() (map[testID]int64, int64) {
	seenIDs := make(map[testID]bool)
	issuers := make(map[testID]int64)
	var totalWeight int64

	for i := 0; i < 10; i++ {
		id := testID([]byte{tpkg.RandByte()})
		if _, exist := seenIDs[id]; exist {
			i--
			continue
		}
		issuers[id] = int64(i)
		totalWeight += int64(i)
		seenIDs[id] = true
	}

	return issuers, totalWeight
}

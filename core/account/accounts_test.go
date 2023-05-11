package account_test

import (
	"testing"

	"github.com/iotaledger/hive.go/core/account"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/stretchr/testify/require"
)

// Define a test struct to hold test data.
type accountTestCase struct {
	id          testID
	weight      int64
	expectedGet int64
	expectedMap map[testID]int64
}

func getTestCases() []accountTestCase {
	return []accountTestCase{
		{
			id:          testID([]byte{1}),
			weight:      0,
			expectedGet: 0,
			expectedMap: map[testID]int64{
				testID([]byte{1}): 0,
			},
		},
		{
			id:          testID([]byte{2}),
			weight:      10,
			expectedGet: 10,
			expectedMap: map[testID]int64{
				testID([]byte{1}): 0,
				testID([]byte{2}): 10,
			},
		},
	}
}

// Define a test function to test the Account type.
func TestAccount(t *testing.T) {
	store := mapdb.NewMapDB()

	// Define your test data.
	testCases := getTestCases()

	// Create a new Accounts instance.
	accounts := account.NewAccounts[testID](store)

	// Loop over each test case and run the tests.
	for _, tc := range testCases {
		// Set the account weight.
		accounts.Set(tc.id, tc.weight)

		// Test the Get() function.
		weight, exists := accounts.Get(tc.id)
		require.Truef(t, exists, "account %v not exist", tc.id)
		require.Equal(t, tc.weight, weight, "account weight does not match from Get")

		// Test the Map() function.
		actualMap, err := accounts.Map()
		require.NoError(t, err, "Map() returned an error")

		for id, weight := range tc.expectedMap {
			actualWeight, ok := actualMap[id]
			require.True(t, ok, "account not exists from Map()")
			require.Equal(t, weight, actualWeight, "account weight is different from Map()")
		}
	}
}

type testID [1]byte

func (t testID) Bytes() ([]byte, error) {
	return t[:], nil
}

func (t *testID) FromBytes(b []byte) (int, error) {
	copy(t[:], b)
	return len(t), nil
}

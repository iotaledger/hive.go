package typedkey

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
)

func Test(t *testing.T) {
	// create a new mapdb instance
	storage := mapdb.NewMapDB()

	// create new StorableCommitment instance
	storableCommitment := NewGenericType[Commitment](storage, 1)
	storableCommitment.Set(Commitment{
		Index:            1,
		PrevID:           [32]byte{1, 2, 3},
		RootsID:          [32]byte{4, 5, 6},
		CumulativeWeight: 789,
	})

	// create new StorableCommitment instance with the same storage and type
	storableCommitment = NewGenericType[Commitment](storage, 1)

	// load the stored commitment
	require.Equal(t, Commitment{
		Index:            1,
		PrevID:           [32]byte{1, 2, 3},
		RootsID:          [32]byte{4, 5, 6},
		CumulativeWeight: 789,
	}, storableCommitment.Get())
}

// Commitment is a somewhat complex type used to test the storable Type.
type Commitment struct {
	Index            int64
	PrevID           [32]byte
	RootsID          [32]byte
	CumulativeWeight int64
}

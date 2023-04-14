package account

import (
	"github.com/iotaledger/hive.go/core/index"
	"github.com/iotaledger/hive.go/crypto/identity"
)

// AccountsUpdateBatch is a batch of weight diffs that can be applied to a Weights instance.
type AccountsUpdateBatch[I index.Type] struct {
	targetSlot I
	diffs      map[identity.ID]int64
	totalDiff  int64
}

// NewAccountsUpdateBatch creates a new WeightsBatch instance.
func NewAccountsUpdateBatch[I index.Type](targetSlot I) *AccountsUpdateBatch[I] {
	return &AccountsUpdateBatch[I]{
		targetSlot: targetSlot,
		diffs:      make(map[identity.ID]int64),
	}
}

// Update updates the weight diff of the given identity.
func (w *AccountsUpdateBatch[I]) Update(id identity.ID, diff int64) {
	if w.diffs[id] += diff; w.diffs[id] == 0 {
		delete(w.diffs, id)
	}

	w.totalDiff += diff
}

func (w *AccountsUpdateBatch[I]) Get(id identity.ID) (diff int64) {
	return w.diffs[id]
}

// TargetIndex returns the slot that the batch is targeting.
func (w *AccountsUpdateBatch[I]) TargetIndex() (targetSlot I) {
	return w.targetSlot
}

// ForEach iterates over all weight diffs in the batch.
func (w *AccountsUpdateBatch[I]) ForEach(consumer func(id identity.ID, diff int64)) {
	for id, diff := range w.diffs {
		consumer(id, diff)
	}
}

// TotalDiff returns the total weight diff of the batch.
func (w *AccountsUpdateBatch[I]) TotalDiff() (totalDiff int64) {
	return w.totalDiff
}

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/generics/lo"
	"github.com/iotaledger/hive.go/types"
)

type SigLockedSingleOutput struct {
	Model[types.Identifier, sigLockedSingleOutput]
}

type sigLockedSingleOutput struct {
	Balance uint64 `serix:"0"`
	Address uint64 `serix:"1"`
}

func NewSigLockedSingleOutput(balance uint64, address uint64) *SigLockedSingleOutput {
	return &SigLockedSingleOutput{NewModel[types.Identifier](sigLockedSingleOutput{
		Balance: balance,
		Address: address,
	}, func(model *sigLockedSingleOutput) types.Identifier {
		return types.Identifier{1}
	})}
}

func (s *SigLockedSingleOutput) Balance() uint64 {
	s.RLock()
	defer s.RUnlock()

	return s.m.Balance
}

func (s *SigLockedSingleOutput) Address() uint64 {
	s.RLock()
	defer s.RUnlock()

	return s.m.Address
}

func TestSth(t *testing.T) {
	source := NewSigLockedSingleOutput(1337, 2)

	restored := lo.NewInstance[SigLockedSingleOutput]()
	restoredObject, err := restored.FromObjectStorage(source.ObjectStorageKey(), source.ObjectStorageValue())
	assert.NoError(t, err)

	assert.Equal(t, source.ID(), restored.ID())
	assert.Equal(t, source.Address(), restored.Address())
	assert.Equal(t, source.Balance(), restored.Balance())
	assert.Equal(t, source.ObjectStorageKey(), restoredObject.ObjectStorageKey())
	assert.Equal(t, source.ObjectStorageValue(), restoredObject.ObjectStorageValue())
}

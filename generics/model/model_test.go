package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/types"
)

func TestModel(t *testing.T) {
	source := NewSigLockedSingleOutput(1337, 2)

	restored := new(SigLockedSingleOutput)
	assert.NoError(t, restored.FromObjectStorage(source.ObjectStorageKey(), source.ObjectStorageValue()))

	assert.Equal(t, source.ID(), restored.ID())
	assert.Equal(t, source.Address(), restored.Address())
	assert.Equal(t, source.Balance(), restored.Balance())
	assert.Equal(t, source.ObjectStorageKey(), restored.ObjectStorageKey())
	assert.Equal(t, source.ObjectStorageValue(), restored.ObjectStorageValue())

	fmt.Println(restored)
}

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

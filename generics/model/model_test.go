package model

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/generics/lo"
	"github.com/iotaledger/hive.go/serix"
	"github.com/iotaledger/hive.go/types"
)

func TestModel(t *testing.T) {
	source := NewSigLockedSingleOutputModel(1337, 2)

	restored := new(SigLockedSingleOutputModel)

	var sth SigLockedSingleOutputModel
	_, err := serix.DefaultAPI.Decode(context.Background(), lo.PanicOnErr(source.Bytes()), &sth, serix.WithValidation())
	assert.NoError(t, err)

	assert.NoError(t, restored.FromBytes(lo.PanicOnErr(source.Bytes())))

	assert.Equal(t, source.Address(), restored.Address())
	assert.Equal(t, source.Balance(), restored.Balance())

	fmt.Println(source, restored)
}

func TestStorable(t *testing.T) {
	source := NewSigLockedSingleOutputStorable(1337, 2)
	source.SetID(types.NewIdentifier([]byte("sigLockedSingleOutput")))

	restored := new(SigLockedSingleOutputStorable)
	assert.NoError(t, restored.FromObjectStorage(source.ObjectStorageKey(), source.ObjectStorageValue()))

	assert.Equal(t, source.ID(), restored.ID())
	assert.Equal(t, source.Address(), restored.Address())
	assert.Equal(t, source.Balance(), restored.Balance())
	assert.Equal(t, source.ObjectStorageKey(), restored.ObjectStorageKey())
	assert.Equal(t, source.ObjectStorageValue(), restored.ObjectStorageValue())

	fmt.Println(restored)
}

type SigLockedSingleOutputModel struct {
	Model[SigLockedSingleOutputModel, *SigLockedSingleOutputModel, sigLockedSingleOutput] `serix:"0"`
}

func NewSigLockedSingleOutputModel(balance uint64, address uint64) *SigLockedSingleOutputModel {
	return New[SigLockedSingleOutputModel](&sigLockedSingleOutput{
		Balance: balance,
		Address: address,
	})
}

func (s *SigLockedSingleOutputModel) Balance() uint64 {
	s.RLock()
	defer s.RUnlock()

	return s.M.Balance
}

func (s *SigLockedSingleOutputModel) Address() uint64 {
	s.RLock()
	defer s.RUnlock()

	return s.M.Address
}

type sigLockedSingleOutput struct {
	Balance uint64 `serix:"0"`
	Address uint64 `serix:"1"`
}

type SigLockedSingleOutputStorable struct {
	Storable[types.Identifier, SigLockedSingleOutputStorable, *SigLockedSingleOutputStorable, sigLockedSingleOutput] `serix:"0"`
}

func NewSigLockedSingleOutputStorable(balance uint64, address uint64) *SigLockedSingleOutputStorable {
	return NewStorable[types.Identifier, SigLockedSingleOutputStorable](&sigLockedSingleOutput{
		Balance: balance,
		Address: address,
	})
}

func (s *SigLockedSingleOutputStorable) Balance() uint64 {
	s.RLock()
	defer s.RUnlock()

	return s.M.Balance
}

func (s *SigLockedSingleOutputStorable) Address() uint64 {
	s.RLock()
	defer s.RUnlock()

	return s.M.Address
}

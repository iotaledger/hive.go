//nolint:tagliatelle // we don't care about these linters in test cases
package model

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/izuc/zipp.foundation/lo"
	"github.com/izuc/zipp.foundation/serializer/v2/serix"
)

func TestModel(t *testing.T) {
	source := NewSigLockedSingleOutputModel(1337, 2)

	restored := new(SigLockedSingleOutputModel)

	var sth SigLockedSingleOutputModel
	_, err := serix.DefaultAPI.Decode(context.Background(), lo.PanicOnErr(source.Bytes()), &sth, serix.WithValidation())
	assert.NoError(t, err)

	_, err = restored.FromBytes(lo.PanicOnErr(source.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, source.Address(), restored.Address())
	assert.Equal(t, source.Balance(), restored.Balance())

	fmt.Println(source, restored)
}

type sigLockedSingleOutput struct {
	Balance uint64 `serix:"0"`
	Address uint64 `serix:"1"`
}

type SigLockedSingleOutputModel struct {
	Mutable[SigLockedSingleOutputModel, *SigLockedSingleOutputModel, sigLockedSingleOutput] `serix:"0"`
}

func NewSigLockedSingleOutputModel(balance uint64, address uint64) *SigLockedSingleOutputModel {
	return NewMutable[SigLockedSingleOutputModel](&sigLockedSingleOutput{
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

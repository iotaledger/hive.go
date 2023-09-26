//nolint:tagliatelle // we don't care about these linters in test cases
package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/izuc/zipp.foundation/ds/types"
	"github.com/izuc/zipp.foundation/lo"
)

func TestStorable(t *testing.T) {
	source := NewSigLockedSingleOutputStorable(1337, 2)
	source.SetID(types.NewIdentifier([]byte("sigLockedSingleOutput")))

	fmt.Println(source)

	restored := new(SigLockedSingleOutputStorable)
	assert.NoError(t, restored.FromObjectStorage(source.ObjectStorageKey(), source.ObjectStorageValue()))

	assert.Equal(t, source.ID(), restored.ID())
	assert.Equal(t, source.Address(), restored.Address())
	assert.Equal(t, source.Balance(), restored.Balance())
	assert.Equal(t, source.ObjectStorageKey(), restored.ObjectStorageKey())
	assert.Equal(t, source.ObjectStorageValue(), restored.ObjectStorageValue())
	assert.Equal(t, lo.PanicOnErr(source.Bytes()), lo.PanicOnErr(restored.Bytes()))

	restored = new(SigLockedSingleOutputStorable)
	_, err := restored.FromBytes(lo.PanicOnErr(source.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, source.Address(), restored.Address())
	assert.Equal(t, source.Balance(), restored.Balance())
	assert.Equal(t, source.ObjectStorageValue(), restored.ObjectStorageValue())
	assert.Equal(t, lo.PanicOnErr(source.Bytes()), lo.PanicOnErr(restored.Bytes()))
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

func TestReferenceModel(t *testing.T) {
	source := NewChildBranch(types.NewIdentifier([]byte("parent")), types.NewIdentifier([]byte("child")))

	restored := new(ChildBranch)
	assert.NoError(t, restored.FromObjectStorage(source.ObjectStorageKey(), source.ObjectStorageValue()))

	assert.Equal(t, source.ParentBranchID(), restored.ParentBranchID())
	assert.Equal(t, source.ChildBranchID(), restored.ChildBranchID())
	assert.Equal(t, source.ObjectStorageKey(), restored.ObjectStorageKey())
	assert.Equal(t, source.ObjectStorageValue(), restored.ObjectStorageValue())
}

// ChildBranch represents the reference between a Conflict and its children.
type ChildBranch struct {
	StorableReference[ChildBranch, *ChildBranch, types.Identifier, types.Identifier] `serix:"0"`
}

// NewChildBranch return a new ChildBranch reference from the named parent to the named child.
func NewChildBranch(parentBranchID, childBranchID types.Identifier) *ChildBranch {
	return NewStorableReference[ChildBranch](parentBranchID, childBranchID)
}

// ParentBranchID returns the identifier of the parent Conflict.
func (c *ChildBranch) ParentBranchID() (parentBranchID types.Identifier) {
	return c.SourceID()
}

// ChildBranchID returns the identifier of the child Conflict.
func (c *ChildBranch) ChildBranchID() (childBranchID types.Identifier) {
	return c.TargetID()
}

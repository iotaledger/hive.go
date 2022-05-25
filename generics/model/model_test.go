package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/types"
)

func TestModel(t *testing.T) {
	source := NewSigLockedSingleOutput(1337, 2)
	source.SetID(types.NewIdentifier([]byte("sigLockedSingleOutput")))

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

	return s.M.Balance
}

func (s *SigLockedSingleOutput) Address() uint64 {
	s.RLock()
	defer s.RUnlock()

	return s.M.Address
}

// region ReferenceModel ///////////////////////////////////////////////////////////////////////////////////////////////

func TestReferenceModel(t *testing.T) {
	source := NewChildBranch[types.Identifier](types.NewIdentifier([]byte("parent")), types.NewIdentifier([]byte("child")))

	restored := new(ChildBranch[types.Identifier])
	assert.NoError(t, restored.FromObjectStorage(source.ObjectStorageKey(), source.ObjectStorageValue()))

	assert.Equal(t, source.ParentBranchID(), restored.ParentBranchID())
	assert.Equal(t, source.ChildBranchID(), restored.ChildBranchID())
	assert.Equal(t, source.ObjectStorageKey(), restored.ObjectStorageKey())
	assert.Equal(t, source.ObjectStorageValue(), restored.ObjectStorageValue())
}

// ChildBranch represents the reference between a Conflict and its children.
type ChildBranch[ConflictID comparable] struct {
	ReferenceModel[ConflictID, ConflictID]
}

// NewChildBranch return a new ChildBranch reference from the named parent to the named child.
func NewChildBranch[ConflictID comparable](parentBranchID, childBranchID ConflictID) (new *ChildBranch[ConflictID]) {
	new = &ChildBranch[ConflictID]{NewReferenceModel[ConflictID, ConflictID](parentBranchID, childBranchID)}

	return new
}

// ParentBranchID returns the identifier of the parent Conflict.
func (c *ChildBranch[ConflictID]) ParentBranchID() (parentBranchID ConflictID) {
	return c.SourceID
}

// ChildBranchID returns the identifier of the child Conflict.
func (c *ChildBranch[ConflictID]) ChildBranchID() (childBranchID ConflictID) {
	return c.TargetID
}

// endregion ////////////////////////////////////////////////////////////////////////////////////////////////////////////

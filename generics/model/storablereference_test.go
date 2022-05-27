package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/types"
)

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
	StorableReference[ConflictID, ConflictID]
}

// NewChildBranch return a new ChildBranch reference from the named parent to the named child.
func NewChildBranch[ConflictID comparable](parentBranchID, childBranchID ConflictID) (new *ChildBranch[ConflictID]) {
	new = &ChildBranch[ConflictID]{NewStorableReference[ConflictID, ConflictID](parentBranchID, childBranchID)}

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

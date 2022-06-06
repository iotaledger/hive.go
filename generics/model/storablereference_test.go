package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/types"
)

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
func NewChildBranch(parentBranchID, childBranchID types.Identifier) (new *ChildBranch) {
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

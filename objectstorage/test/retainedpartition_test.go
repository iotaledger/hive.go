package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/objectstorage"
)

func TestRetainedPartition(t *testing.T) {
	retainedPartition := objectstorage.NewRetainedPartition()

	assert.Equal(t, false, retainedPartition.IsRetained([]string{"A"}))
	assert.Equal(t, false, retainedPartition.IsRetained([]string{"A", "B"}))

	retainedPartition.Retain([]string{"A", "B"})

	assert.Equal(t, true, retainedPartition.IsRetained([]string{"A"}))
	assert.Equal(t, true, retainedPartition.IsRetained([]string{"A", "B"}))
	assert.Equal(t, false, retainedPartition.IsRetained([]string{"A", "B", "C"}))

	retainedPartition.Release([]string{"A", "B"})

	assert.Equal(t, false, retainedPartition.IsRetained([]string{"A"}))
	assert.Equal(t, false, retainedPartition.IsRetained([]string{"A", "B"}))
}

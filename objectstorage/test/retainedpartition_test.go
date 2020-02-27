package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/objectstorage"
)

func TestRetainedPartition(t *testing.T) {
	retainedPartition := objectstorage.NewPartitionsManager()

	assert.Equal(t, false, retainedPartition.IsRetained([]string{"A"}))
	assert.Equal(t, false, retainedPartition.IsRetained([]string{"A", "B"}))

	retainedPartition.Retain([]string{"A", "B"})

	assert.Equal(t, true, retainedPartition.IsRetained([]string{"A"}))
	assert.Equal(t, true, retainedPartition.IsRetained([]string{"A", "B"}))
	assert.Equal(t, false, retainedPartition.IsRetained([]string{"A", "B", "C"}))

	retainedPartition.Retain([]string{"A"})

	retainedPartition.Release([]string{"A", "B"})

	assert.Equal(t, true, retainedPartition.IsRetained([]string{"A"}))
	assert.Equal(t, false, retainedPartition.IsRetained([]string{"A", "B"}))

	retainedPartition.Release([]string{"A"})
}

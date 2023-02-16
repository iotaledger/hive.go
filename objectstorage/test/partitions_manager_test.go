package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/objectstorage"
)

func TestPartitionsManager(t *testing.T) {
	partitionsManager := objectstorage.NewPartitionsManager()

	assert.Equal(t, false, partitionsManager.IsRetained([]string{"A"}))
	assert.Equal(t, false, partitionsManager.IsRetained([]string{"A", "B"}))

	partitionsManager.Retain([]string{"A", "B"})

	assert.Equal(t, true, partitionsManager.IsRetained([]string{"A"}))
	assert.Equal(t, true, partitionsManager.IsRetained([]string{"A", "B"}))
	assert.Equal(t, false, partitionsManager.IsRetained([]string{"A", "B", "C"}))

	partitionsManager.Retain([]string{"A"})

	partitionsManager.Release([]string{"A", "B"})

	assert.Equal(t, true, partitionsManager.IsRetained([]string{"A"}))
	assert.Equal(t, false, partitionsManager.IsRetained([]string{"A", "B"}))

	partitionsManager.Release([]string{"A"})
}

package kvstore_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
)

var sequenceKey = []byte("test_sequence")

const sequenceInterval = 100

func TestNewSequence(t *testing.T) {
	store := mapdb.NewMapDB()
	s1, err := kvstore.NewSequence(store, sequenceKey, 1)
	assert.NotNil(t, s1)
	assert.NoError(t, err)

	newKey, err := s1.Next()
	require.NoError(t, err)
	assert.Equal(t, uint64(0), newKey)

	// key should exists in the store
	_, err = store.Get(sequenceKey)
	require.NoError(t, err)
}

func TestSequence_Next(t *testing.T) {
	store := mapdb.NewMapDB()
	s1, err := kvstore.NewSequence(store, sequenceKey, sequenceInterval)
	require.NoError(t, err)

	for i := 0; i < sequenceInterval; i++ {
		next, err := s1.Next()
		require.NoError(t, err)
		assert.EqualValues(t, i, next)
	}

	s2, err := kvstore.NewSequence(store, sequenceKey, sequenceInterval)
	require.NoError(t, err)

	next2, err := s2.Next()
	require.NoError(t, err)
	assert.EqualValues(t, sequenceInterval, next2)

	s3, err := kvstore.NewSequence(store, sequenceKey, sequenceInterval)
	require.NoError(t, err)

	next3, err := s3.Next()
	require.NoError(t, err)
	assert.EqualValues(t, 2*sequenceInterval, next3)
}

func TestSequence_Release(t *testing.T) {
	store := mapdb.NewMapDB()
	s1, err := kvstore.NewSequence(store, sequenceKey, math.MaxUint64)
	require.NoError(t, err)

	first, err := s1.Next()
	require.NoError(t, err)

	require.NoError(t, s1.Release())

	s2, err := kvstore.NewSequence(store, sequenceKey, 1)
	require.NoError(t, err)

	second, err := s2.Next()
	require.NoError(t, err)

	assert.Equal(t, first+1, second)
}

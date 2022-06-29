package shrinkingmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMapShrinking(t *testing.T) {
	shrink := New[int, int](2)
	for i := 0; i < 100; i++ {
		shrink.Set(i, i)
	}

	assert.Equal(t, 0, shrink.deletedKeys)

	for i := 0; i < 75; i++ {
		assert.Equal(t, i, shrink.deletedKeys)
		shrink.Delete(i)
	}

	assert.Equal(t, 0, shrink.deletedKeys)
	shrink.Delete(99)
	assert.Equal(t, 1, shrink.deletedKeys)
	shrink.Shrink()
	assert.Equal(t, 0, shrink.deletedKeys)
}

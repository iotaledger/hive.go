package byteutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcatBytes(t *testing.T) {
	assert.Equal(t, []byte{1, 2, 3, 4, 5, 7, 8}, ConcatBytes([]byte{1, 2, 3}, []byte{4, 5}, []byte{7, 8}))
	assert.Equal(t, []byte{1, 2, 3, 4, 5, 7, 8}, ConcatBytes([]byte{1, 2, 3, 4, 5, 7, 8}))
}

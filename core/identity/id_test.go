package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestID_Base58EncodeAndDecode(t *testing.T) {
	id, err := RandomIDInsecure()
	require.NoError(t, err)
	encoded := id.EncodeBase58()
	decoded, err := DecodeIDBase58(encoded)
	require.NoError(t, err)
	assert.Equal(t, id, decoded)
}

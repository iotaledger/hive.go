package storable_test

import (
	"path/filepath"
	"testing"

	"github.com/iotaledger/hive.go/core/storable"
	"github.com/stretchr/testify/require"
)

func TestByteSlice(t *testing.T) {
	// Create a new ByteSlice and write a couple of entries
	filePath := filepath.Join(t.TempDir(), "test.bin")
	bs, err := storable.NewByteSlice(filePath, 4)
	require.NoError(t, err, "Failed to create ByteSlice")

	err = bs.Set(0, []byte{0x00, 0x01, 0x02, 0x03})
	require.NoError(t, err, "Failed to set entry")

	err = bs.Set(1, []byte{0x04, 0x05, 0x06, 0x07})
	require.NoError(t, err, "Failed to set entry")

	// Read the entries back and compare them
	entry, err := bs.Get(0)
	require.NoError(t, err, "Failed to get entry")
	require.Equal(t, 4, len(entry))
	require.ElementsMatch(t, []byte{0x00, 0x01, 0x02, 0x03}, entry)

	entry, err = bs.Get(1)
	require.NoError(t, err, "Failed to get entry")
	require.Equal(t, 4, len(entry))
	require.ElementsMatch(t, []byte{0x04, 0x05, 0x06, 0x07}, entry)

	// Attempt to read an out-of-bounds entry
	_, err = bs.Get(2)
	require.Error(t, err, "Expected out-of-bounds error")

	// Attempt to write an entry with the wrong length
	err = bs.Set(2, []byte{0x00, 0x01, 0x02})
	require.Error(t, err, "Expected length error")
}

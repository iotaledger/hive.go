package storable_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/storable"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

// region Tests ////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestStruct(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "node.settings")

	settings := NewSettings(filePath)
	require.Equal(t, uint64(123), settings.Number)
	require.Equal(t, filePath, settings.FilePath())

	settings.Number = 3
	size, err := settings.Size()
	require.Error(t, err)
	require.Equal(t, int64(0), size)

	require.NoError(t, settings.ToFile())
	size, err = settings.Size()
	require.NoError(t, err)

	// write to another file path
	filePath1 := filepath.Join(t.TempDir(), "node1.settings")
	require.NoError(t, settings.ToFile(filePath1))

	// restore from file path
	restoredSettings := NewSettings(filePath)
	require.Equal(t, settings.Number, restoredSettings.Number)
	restoredSize, err := restoredSettings.Size()
	require.NoError(t, err)
	require.Equal(t, size, restoredSize)

	// restore from file path 1
	restoredSettings1 := NewSettings(filePath1)
	require.Equal(t, settings.Number, restoredSettings1.Number)
	restoredSize1, err := restoredSettings.Size()
	require.NoError(t, err)
	require.Equal(t, size, restoredSize1)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Settings /////////////////////////////////////////////////////////////////////////////////////////////////////

type Settings struct {
	Number uint64 `serix:"1"`

	storable.Struct[Settings, *Settings]
}

func NewSettings(filePath string) (settings *Settings) {
	return storable.InitStruct(&Settings{
		Number: 123,
	}, filePath)
}

func (t *Settings) FromBytes(bytes []byte) (int, error) {
	return serix.DefaultAPI.Decode(context.Background(), bytes, t)
}

func (t *Settings) Bytes() ([]byte, error) {
	return serix.DefaultAPI.Encode(context.Background(), *t)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

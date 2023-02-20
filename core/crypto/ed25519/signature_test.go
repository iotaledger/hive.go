package ed25519

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/lo"
)

func TestSignatureFromBytesTooShort(t *testing.T) {
	bytes := make([]byte, 10)
	_, _, err := SignatureFromBytes(bytes)
	require.EqualError(t, err, ErrNotEnoughBytes.Error())
}

func TestSignatureFromBytes(t *testing.T) {
	bytes := make([]byte, 128)
	copy(bytes, "PublicKeyAndSomeOtherDataAndSomeOtherDataAndSomeOtherDataPrivateKeyAndSomeOtherData")

	signature, consumedBytes, err := SignatureFromBytes(bytes)

	require.Equal(t, lo.PanicOnErr(signature.Bytes()), bytes[:SignatureSize])
	require.NoError(t, err)
	require.Equal(t, consumedBytes, SignatureSize)
}

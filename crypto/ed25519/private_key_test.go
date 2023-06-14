package ed25519

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/lo"
)

func TestPrivateKeyFromBytesTooShort(t *testing.T) {
	bytes := make([]byte, 10)
	_, _, err := PrivateKeyFromBytes(bytes)
	require.EqualError(t, err, ErrNotEnoughBytes.Error())
}

func TestPrivateKeyFromBytes(t *testing.T) {
	bytes := make([]byte, 128)
	copy(bytes, "PrivateKeyAndSomeOtherDataAndSomeOtherDataAndSomeOtherDataPrivateKeyAndSomeOtherData")

	privateKey, consumedBytes, err := PrivateKeyFromBytes(bytes)

	require.Equal(t, lo.PanicOnErr(privateKey.Bytes()), bytes[:PrivateKeySize])
	require.NoError(t, err)
	require.Equal(t, consumedBytes, PrivateKeySize)
}

func TestPrivateKeyFromSeed(t *testing.T) {
	seed := make([]byte, SeedSize)
	copy(seed, "MySeed")

	privateKey := PrivateKeyFromSeed(seed)

	require.EqualValues(t, lo.PanicOnErr(privateKey.Bytes()), ed25519.NewKeyFromSeed(seed))
}

func TestPrivateKey_Sign(t *testing.T) {
	privateKey, err := GeneratePrivateKey()
	require.NoError(t, err)

	data := []byte("DataToSign")
	sig := privateKey.Sign(data)

	require.Equal(t, lo.PanicOnErr(sig.Bytes()), ed25519.Sign(lo.PanicOnErr(privateKey.Bytes()), data))
}

func TestPrivateKey_Public(t *testing.T) {
	publicKey, privateKey, err := GenerateKey()
	require.NoError(t, err)

	require.Equal(t, privateKey.Public(), publicKey)
}

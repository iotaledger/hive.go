package ed25519

import (
	"testing"

	"github.com/oasisprotocol/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrivateKeyFromBytesTooShort(t *testing.T) {
	bytes := make([]byte, 10)
	_, err, _ := PrivateKeyFromBytes(bytes)
	assert.EqualError(t, err, "bytes too short")
}

func TestPrivateKeyFromBytes(t *testing.T) {
	bytes := make([]byte, 128)
	copy(bytes, "PrivateKeyAndSomeOtherDataAndSomeOtherDataAndSomeOtherDataPrivateKeyAndSomeOtherData")

	privateKey, err, consumedBytes := PrivateKeyFromBytes(bytes)

	assert.Equal(t, privateKey.Bytes(), bytes[:PrivateKeySize])
	assert.NoError(t, err)
	assert.Equal(t, consumedBytes, PrivateKeySize)
}

func TestPrivateKeyFromSeed(t *testing.T) {
	seed := make([]byte, SeedSize)
	copy(seed, "MySeed")

	privateKey := PrivateKeyFromSeed(seed)

	assert.EqualValues(t, privateKey.Bytes(), ed25519.NewKeyFromSeed(seed))
}

func TestPrivateKey_Sign(t *testing.T) {
	privateKey, err := GeneratePrivateKey()
	require.NoError(t, err)

	data := []byte("DataToSign")
	sig := privateKey.Sign(data)

	assert.Equal(t, sig.Bytes(), ed25519.Sign(privateKey.Bytes(), data))
}

func TestPrivateKey_Public(t *testing.T) {
	publicKey, privateKey, err := GenerateKey()
	require.NoError(t, err)

	assert.Equal(t, privateKey.Public(), publicKey)
}

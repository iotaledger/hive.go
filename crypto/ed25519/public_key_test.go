package ed25519

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublicKeyFromBytesTooShort(t *testing.T) {
	bytes := make([]byte, 10)
	_, _, err := PublicKeyFromBytes(bytes)
	assert.EqualError(t, err, "bytes too short")
}

func TestPublicKeyFromBytes(t *testing.T) {
	bytes := make([]byte, 128)
	copy(bytes, "PublicKeyAndSomeOtherDataAndSomeOtherDataAndSomeOtherDataPrivateKeyAndSomeOtherData")

	publicKey, consumedBytes, err := PublicKeyFromBytes(bytes)

	assert.Equal(t, publicKey.Bytes(), bytes[:PublicKeySize])
	assert.NoError(t, err)
	assert.Equal(t, consumedBytes, PublicKeySize)
}

func TestPublicKey_VerifySignature(t *testing.T) {
	publicKey, privateKey, err := GenerateKey()
	require.NoError(t, err)

	data := []byte("DataToSign")
	sig := privateKey.Sign(data)

	assert.True(t, publicKey.VerifySignature(data, sig))
}

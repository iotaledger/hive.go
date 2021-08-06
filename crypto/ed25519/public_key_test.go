package ed25519

import (
	"encoding/json"
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

func TestPublicKey_MarshalJSON(t *testing.T) {
	pk, err := PublicKeyFromString("CHfU1NUf6ZvUKDQHTG2df53GR7CvuMFtyt7YymJ6DwS3")
	require.NoError(t, err)
	b, err := json.Marshal(pk)
	require.NoError(t, err)
	got := string(b)
	assert.Equal(t, `"CHfU1NUf6ZvUKDQHTG2df53GR7CvuMFtyt7YymJ6DwS3"`, got)
}

func TestPublicKey_UnmarshalJSON(t *testing.T) {
	jsonData := `"CHfU1NUf6ZvUKDQHTG2df53GR7CvuMFtyt7YymJ6DwS3"`
	var got PublicKey
	err := json.Unmarshal([]byte(jsonData), &got)
	require.NoError(t, err)

	expected, err := PublicKeyFromString("CHfU1NUf6ZvUKDQHTG2df53GR7CvuMFtyt7YymJ6DwS3")
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

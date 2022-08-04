package ed25519

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignatureFromBytesTooShort(t *testing.T) {
	bytes := make([]byte, 10)
	_, _, err := SignatureFromBytes(bytes)
	assert.EqualError(t, err, "bytes too short")
}

func TestSignatureFromBytes(t *testing.T) {
	bytes := make([]byte, 128)
	copy(bytes, "PublicKeyAndSomeOtherDataAndSomeOtherDataAndSomeOtherDataPrivateKeyAndSomeOtherData")

	signature, consumedBytes, err := SignatureFromBytes(bytes)

	assert.Equal(t, signature.Bytes(), bytes[:SignatureSize])
	assert.NoError(t, err)
	assert.Equal(t, consumedBytes, SignatureSize)
}

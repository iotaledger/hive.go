package bls

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dataToSign = []byte("Hello BLS Test!")

func TestAggregateSignatures(t *testing.T) {
	signatureCount := 20

	signatures := make([]SignatureWithPublicKey, signatureCount)
	privateKeys := make([]PrivateKey, signatureCount)

	for i := range signatures {
		privateKeys[i] = PrivateKeyFromRandomness()
		signature, err := privateKeys[i].Sign(dataToSign)
		require.NoError(t, err)
		signatures[i] = signature
	}

	// aggregate 2 signatures
	a01, err := AggregateSignatures(signatures[0], signatures[1])
	require.NoError(t, err)
	assert.True(t, a01.IsValid(dataToSign))

	// aggregate N signatures
	aN, err := AggregateSignatures(signatures...)
	require.NoError(t, err)
	assert.True(t, aN.IsValid(dataToSign))
}

func TestSingleSignature(t *testing.T) {
	privateKey := PrivateKeyFromRandomness()

	signature, err := privateKey.Sign(dataToSign)
	require.NoError(t, err)
	assert.True(t, signature.IsValid(dataToSign))
	assert.False(t, signature.IsValid([]byte("some other data")))
}

func TestMarshalPublicKey(t *testing.T) {
	privateKey := PrivateKeyFromRandomness()
	pubKey := privateKey.PublicKey()

	pubKeyBytes := pubKey.Bytes()
	pubKeyBack, _, err := PublicKeyFromBytes(pubKeyBytes)
	require.NoError(t, err)
	require.EqualValues(t, pubKeyBytes, pubKeyBack.Bytes())
}

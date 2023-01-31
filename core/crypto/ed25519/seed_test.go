package ed25519

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSeed(t *testing.T) {
	randomSeed := NewSeed()

	// check if NewSeed generates different seeds in consequent calls.
	require.NotEqual(t, randomSeed, NewSeed())

	// check if the key derivation logic is deterministic
	require.Equal(t, randomSeed.KeyPair(0), randomSeed.KeyPair(0))
	require.Equal(t, randomSeed.KeyPair(1337), randomSeed.KeyPair(1337))

	// check if the generated KeyPairs can sign and verify
	someKeyPair := randomSeed.KeyPair(7)
	signature := someKeyPair.PrivateKey.Sign([]byte{1, 3, 3, 8})
	require.True(t, someKeyPair.PublicKey.VerifySignature([]byte{1, 3, 3, 8}, signature))
}

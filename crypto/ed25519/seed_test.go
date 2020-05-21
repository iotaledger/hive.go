package ed25519

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSeed(t *testing.T) {
	randomSeed := NewSeed()

	// check if NewSeed generates different seeds in consequent calls.
	assert.NotEqual(t, randomSeed, NewSeed())

	// check if the key derivation logic is deterministic
	assert.Equal(t, randomSeed.KeyPair(0), randomSeed.KeyPair(0))
	assert.Equal(t, randomSeed.KeyPair(1337), randomSeed.KeyPair(1337))

	// check if the generated KeyPairs can sign and verify
	someKeyPair := randomSeed.KeyPair(7)
	signature := someKeyPair.PrivateKey.Sign([]byte{1, 3, 3, 8})
	assert.True(t, someKeyPair.PublicKey.VerifySignature([]byte{1, 3, 3, 8}, signature))
}

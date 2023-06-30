package types

import (
	"crypto/rand"
	"sync"

	"github.com/mr-tron/base58"
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
)

func init() {

}

var (
	// ErrBase58DecodeFailed is returned if a base58 encoded string can not be decoded.
	ErrBase58DecodeFailed = ierrors.New("failed to decode base58 encoded string")
)

// region Identifier ///////////////////////////////////////////////////////////////////////////////////////////////////

// Identifier is a 32 byte hash value that can be used to uniquely identify some blob of data.
type Identifier [IdentifierLength]byte

// NewIdentifier returns a new Identifier for the given data.
func NewIdentifier(data []byte) Identifier {
	return blake2b.Sum256(data)
}

// FromRandomness generates a random Identifier.
func (t *Identifier) FromRandomness() (err error) {
	_, err = rand.Read((*t)[:])

	return
}

// FromBase58 un-serializes an Identifier from a base58 encoded string.
func (t *Identifier) FromBase58(base58String string) (err error) {
	decodedBytes, err := base58.Decode(base58String)
	if err != nil {
		return ierrors.Wrapf(ErrBase58DecodeFailed, "error while decoding base58 encoded Identifier (%v)", err)
	}

	if _, err = t.Decode(decodedBytes); err != nil {
		return ierrors.Wrap(err, "failed to parse Identifier from bytes")
	}

	return nil
}

// RegisterAlias allows to register a human-readable alias for the Identifier which will be used as a replacement for
// the String method.
func (t Identifier) RegisterAlias(alias string) {
	identifierAliasesMutex.Lock()
	defer identifierAliasesMutex.Unlock()

	identifierAliases[t] = alias
}

// Alias returns the human-readable alias of the Identifier (or the base58 encoded bytes of no alias was set).
func (t Identifier) Alias() (alias string) {
	identifierAliasesMutex.RLock()
	defer identifierAliasesMutex.RUnlock()

	if existingAlias, exists := identifierAliases[t]; exists {
		return existingAlias
	}

	return t.Base58()
}

// UnregisterAlias allows to unregister a previously registered alias.
func (t Identifier) UnregisterAlias() {
	identifierAliasesMutex.Lock()
	defer identifierAliasesMutex.Unlock()

	delete(identifierAliases, t)
}

// Decode decodes the Identifier from a sequence of bytes.
func (t *Identifier) Decode(data []byte) (consumed int, err error) {
	if len(data) < IdentifierLength {
		return 0, ierrors.New("not enough data to decode Identifier")
	}
	copy(t[:], data[:IdentifierLength])

	return IdentifierLength, nil
}

// Encode returns a serialized version of the Identifier.
func (t Identifier) Encode() (serialized []byte, err error) {
	return t[:], nil
}

// Bytes returns the raw bytes of the Identifier.
func (t Identifier) Bytes() []byte {
	return t[:]
}

// Base58 returns a base58 encoded version of the Identifier.
func (t Identifier) Base58() (base58Encoded string) {
	return base58.Encode(t[:])
}

// String returns a human-readable version of the Identifier.
func (t Identifier) String() (humanReadable string) {
	return "Identifier(" + t.Alias() + ")"
}

// IdentifierLength contains the byte length of a serialized Identifier.
const IdentifierLength = 32

var (
	// identifierAliases contains a dictionary of identifiers associated to their human-readable alias.
	identifierAliases = make(map[Identifier]string)

	// identifierAliasesMutex is the mutex that is used to synchronize access to the previous map.
	identifierAliasesMutex = sync.RWMutex{}
)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

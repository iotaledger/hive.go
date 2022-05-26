package types

import (
	"crypto/rand"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/mr-tron/base58"
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/cerrors"
)

func init() {

}

// region Identifier ///////////////////////////////////////////////////////////////////////////////////////////////////

// Identifier is a 32 byte hash value that can be used to uniquely identify some blob of data.
type Identifier [IdentifierLength]byte

// NewIdentifier returns a new Identifier for the given data.
func NewIdentifier(data []byte) (new Identifier) {
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
		return errors.Errorf("error while decoding base58 encoded Identifier (%v): %w", err, cerrors.ErrBase58DecodeFailed)
	}

	if _, err = t.Decode(decodedBytes); err != nil {
		return errors.Errorf("failed to parse Identifier from bytes: %w", err)
	}

	return nil
}

// RegisterAlias allows to register a human-readable alias for the Identifier which will be used as a replacement for
// the String method.
func (t Identifier) RegisterAlias(alias string) {
	_identifierAliasesMutex.Lock()
	defer _identifierAliasesMutex.Unlock()

	_identifierAliases[t] = alias
}

// Alias returns the human-readable alias of the Identifier (or the base58 encoded bytes of no alias was set).
func (t Identifier) Alias() (alias string) {
	_identifierAliasesMutex.RLock()
	defer _identifierAliasesMutex.RUnlock()

	if existingAlias, exists := _identifierAliases[t]; exists {
		return existingAlias
	}

	return t.Base58()
}

// UnregisterAlias allows to unregister a previously registered alias.
func (t Identifier) UnregisterAlias() {
	_identifierAliasesMutex.Lock()
	defer _identifierAliasesMutex.Unlock()

	delete(_identifierAliases, t)
}

// Decode decodes the Identifier from a sequence of bytes.
func (t *Identifier) Decode(data []byte) (consumed int, err error) {
	if len(data) < IdentifierLength {
		return 0, errors.New("not enough data to decode Identifier")
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
	// _identifierAliases contains a dictionary of identifiers associated to their human-readable alias.
	_identifierAliases = make(map[Identifier]string)

	// _identifierAliasesMutex is the mutex that is used to synchronize access to the previous map.
	_identifierAliasesMutex = sync.RWMutex{}
)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

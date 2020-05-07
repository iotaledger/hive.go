package objectstorage

import "errors"

var (
	// ErrKeyNotFound is returned when an op. doesn't find the given key.
	ErrKeyNotFound = errors.New("key not found")
)

// IteratorKeyValueConsumerFunc is a consumer function for an iterating function which iterates over keys and values.
// Returning false from this function indicates to abort the iteration.
// They key must not be prefixed with the realm.
type IteratorKeyValueConsumerFunc func(key []byte, value []byte) bool

// IteratorKeyConsumerFunc is a consumer function for an iterating function which iterates only over keys.
// Returning false from this function indicates to abort the iteration.
// They key must not be prefixed with the realm.
type IteratorKeyConsumerFunc func(key []byte) bool

// BatchedMutations represents batched mutations to the storage.
type BatchedMutations interface {
	// Set sets the given key and value.
	Set(realm []byte, key []byte, value []byte) error
	// Delete deletes the entry for the given key.
	Delete(realm []byte, key []byte) error
	// Cancel cancels the batched mutations.
	Cancel()
	// Commit commits/flushes the mutations.
	Commit() error
}

// Storage persists, deletes and retrieves data.
// Realm keys specify namespaces for entities.
type Storage interface {
	// Iterate iterates over all keys or keys with the provided prefix.
	Iterate(realm []byte, prefixes [][]byte, preFetchValues bool, kvConsumerFunc IteratorKeyValueConsumerFunc) error
	// Iterate iterates over all keys with the provided prefix.
	IterateKeys(realm []byte, prefixes [][]byte, consumerFunc IteratorKeyConsumerFunc) error
	// Clear clears the given realm.
	Clear(realm []byte) error
	// Get gets the given key or nil if it doesn't exist or an error if an error occurred.
	Get(realm []byte, key []byte) (value []byte, err error)
	// Set sets the given key and value.
	Set(realm []byte, key []byte, value []byte) error
	// Has checks whether the given key exists.
	Has(realm []byte, key []byte) (bool, error)
	// Delete deletes the entry for the given key.
	Delete(realm []byte, key []byte) error
	// Batched returns a BatchedMutations interface to execute batched mutations.
	Batched() BatchedMutations
}

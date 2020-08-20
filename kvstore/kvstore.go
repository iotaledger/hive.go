package kvstore

import (
	"errors"

	"github.com/iotaledger/hive.go/bitmask"
)

var (
	// ErrKeyNotFound is returned when an op. doesn't find the given key.
	ErrKeyNotFound = errors.New("key not found")

	EmptyPrefix = KeyPrefix{}
)

type Realm = []byte
type KeyPrefix = []byte
type Key = []byte
type Value = []byte

// IteratorKeyValueConsumerFunc is a consumer function for an iterating function which iterates over keys and values.
// They key must not be prefixed with the realm.
// Returning false from this function indicates to abort the iteration.
type IteratorKeyValueConsumerFunc func(key Key, value Value) bool

// IteratorKeyConsumerFunc is a consumer function for an iterating function which iterates only over keys.
// They key must not be prefixed with the realm.
// Returning false from this function indicates to abort the iteration.
type IteratorKeyConsumerFunc func(key Key) bool

// BatchedMutations represents batched mutations to the storage.
type BatchedMutations interface {

	// Set sets the given key and value.
	Set(key Key, value Value) error

	// Delete deletes the entry for the given key.
	Delete(key Key) error

	// Cancel cancels the batched mutations.
	Cancel()

	// Commit commits/flushes the mutations.
	Commit() error
}

type AccessLogCommand byte

type AccessLogFlag = bitmask.BitMask

type AccessLogCallback func(command AccessLogCommand, parameters ...[]byte)

const (
	// IterateCommand represents a call to the Iterate method of the store.
	IterateCommand AccessLogCommand = iota

	// IterateKeysCommand represents a call to the IterateKeys method of the store.
	IterateKeysCommand
	ClearCommand
	GetCommand
	SetCommand
	HasCommand
	DeleteCommand
	DeletePrefixCommand
)

const (
	LogIterateCommand AccessLogFlag = 1 << iota
	LogIterateKeysCommand
	LogClearCommand
	LogGetCommand
	LogSetCommand
	LogHasCommand
	LogDeleteCommand
	LogDeletePrefixCommand

	LogAllCommands = LogIterateCommand | LogIterateKeysCommand | LogClearCommand | LogGetCommand | LogSetCommand | LogHasCommand | LogDeleteCommand | LogDeletePrefixCommand
)

// KVStore persists, deletes and retrieves data.
type KVStore interface {
	// EnableAccessLog configures the store to log all requests by passing the access parameters to the given callback.
	EnableAccessLog(callback AccessLogCallback, accessLogFlag AccessLogFlag)

	// WithRealm is a factory method for using the same underlying storage with a different realm.
	WithRealm(realm Realm) KVStore

	// Realm returns the configured realm.
	Realm() Realm

	// Iterate iterates over all keys and values with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys and values.
	Iterate(prefix KeyPrefix, kvConsumerFunc IteratorKeyValueConsumerFunc) error

	// IterateKeys iterates over all keys with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys.
	IterateKeys(prefix KeyPrefix, consumerFunc IteratorKeyConsumerFunc) error

	// Clear clears the realm.
	Clear() error

	// Get gets the given key or nil if it doesn't exist or an error if an error occurred.
	Get(key Key) (value Value, err error)

	// Set sets the given key and value.
	Set(key Key, value Value) error

	// Has checks whether the given key exists.
	Has(key Key) (bool, error)

	// Delete deletes the entry for the given key.
	Delete(key Key) error

	// DeletePrefix deletes all the entries matching the given key prefix.
	DeletePrefix(prefix KeyPrefix) error

	// Batched returns a BatchedMutations interface to execute batched mutations.
	Batched() BatchedMutations
}

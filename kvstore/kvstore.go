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

// Command is a type that represents a specific method in the KVStore.
type Command = bitmask.BitMask

// AccessCallback is the type of the callback function that can be used to hook the access to the callback.
type AccessCallback func(command Command, parameters ...[]byte)

const (
	// ShutdownCommand represents a call to the Shutdown method of the store.
	ShutdownCommand Command = 0

	// IterateCommand represents a call to the Iterate method of the store.
	IterateCommand Command = 1 << iota

	// IterateKeysCommand represents a call to the IterateKeys method of the store.
	IterateKeysCommand

	// ClearCommand represents a call to the Clear method of the store.
	ClearCommand

	// GetCommand represents a call to the Get method of the store.
	GetCommand

	// SetCommand represents a call to the Set method of the store.
	SetCommand

	// HasCommand represents a call to the Has method of the store.
	HasCommand

	// DeleteCommand represents a call to the Delete method of the store.
	DeleteCommand

	// DeletePrefixCommand represents a call to the DeletePrefix method of the store.
	DeletePrefixCommand

	// AllCommands represents the collection of all commands.
	AllCommands = IterateCommand | IterateKeysCommand | ClearCommand | GetCommand | SetCommand | HasCommand | DeleteCommand | DeletePrefixCommand
)

// CommandNames contains a map from the command to its human readable name.
var CommandNames = map[Command]string{
	ShutdownCommand:     "Shutdown",
	IterateCommand:      "Iterate",
	IterateKeysCommand:  "IterateKeys",
	ClearCommand:        "Clear",
	GetCommand:          "Get",
	SetCommand:          "Set",
	HasCommand:          "Has",
	DeleteCommand:       "Delete",
	DeletePrefixCommand: "DeletePrefix",
}

// KVStore persists, deletes and retrieves data.
type KVStore interface {
	// AccessCallback configures the store to pass all requests to the KVStore to the given callback.
	// This can for example be used for debugging and to examine what the KVStore is doing.
	AccessCallback(callback AccessCallback, commandsFilter ...Command)

	// WithRealm is a factory method for using the same underlying storage with a different realm.
	WithRealm(realm Realm) KVStore

	// Realm returns the configured realm.
	Realm() Realm

	// Shutdown marks the store as shutdown.
	Shutdown()

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

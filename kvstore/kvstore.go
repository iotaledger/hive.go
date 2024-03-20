package kvstore

import (
	"fmt"

	"github.com/iotaledger/hive.go/ierrors"
)

var (
	// ErrKeyNotFound is returned when an op. doesn't find the given key.
	ErrKeyNotFound = ierrors.New("key not found")
	// ErrTypedValueNotChanged is a sentinel error that can be returned by the TypedValue.Compute callback to indicate
	// that the current value should not be changed.
	ErrTypedValueNotChanged = ierrors.New("typed value not changed")
	// ErrStoreClosed is returned when an op accesses the kvstore but it was already closed.
	ErrStoreClosed = ierrors.New("trying to access closed kvstore")

	EmptyPrefix = KeyPrefix{}
)

type Realm = []byte
type KeyPrefix = []byte
type Key = []byte
type Value = []byte

// IterDirection specifies the direction for iterations.
type IterDirection byte

const (
	IterDirectionForward IterDirection = iota
	IterDirectionBackward
)

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

// KVStore persists, deletes and retrieves data.
type KVStore interface {
	// WithRealm is a factory method for using the same underlying storage with a different realm.
	WithRealm(realm Realm) (KVStore, error)

	// WithExtendedRealm is a factory method for using the same underlying storage with an realm appended to existing one.
	WithExtendedRealm(realm Realm) (KVStore, error)

	// Realm returns the configured realm.
	Realm() Realm

	// Iterate iterates over all keys and values with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys and values.
	// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
	Iterate(prefix KeyPrefix, kvConsumerFunc IteratorKeyValueConsumerFunc, direction ...IterDirection) error

	// IterateKeys iterates over all keys with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys.
	// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
	IterateKeys(prefix KeyPrefix, consumerFunc IteratorKeyConsumerFunc, direction ...IterDirection) error

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

	// Flush persists all outstanding write operations to disc.
	Flush() error

	// Close closes the database file handles.
	Close() error

	// Batched returns a BatchedMutations interface to execute batched mutations.
	Batched() (BatchedMutations, error)
}

// GetIterDirection returns the direction to use for an iteration.
// If no direction is given, it defaults to IterDirectionForward.
func GetIterDirection(iterDirection ...IterDirection) IterDirection {
	direction := IterDirectionForward
	if len(iterDirection) > 0 {
		switch iterDirection[0] {
		case IterDirectionForward:
			break
		case IterDirectionBackward:
			direction = iterDirection[0]
		default:
			panic(fmt.Sprintf("unknown iteration direction: %d", iterDirection[0]))
		}
	}

	return direction
}

// Copy copies the content from the source to the target KVStore.
func Copy(source KVStore, target KVStore) error {
	var innerErr error
	if err := source.Iterate(EmptyPrefix, func(key, value Value) bool {
		if err := target.Set(key, value); err != nil {
			innerErr = err
		}

		return innerErr == nil
	}); err != nil {
		return err
	}

	if innerErr != nil {
		return innerErr
	}

	return target.Flush()
}

// CopyBatched copies the content from the source to the target KVStore in batches.
// If batchSize is not specified, everything is copied in a single batch.
func CopyBatched(source KVStore, target KVStore, batchSize ...int) error {
	batchedSize := 0
	if len(batchSize) > 0 {
		batchedSize = batchSize[0]
	}

	// init batched mutation
	currentBatchSize := 0
	batchedMutation, err := target.Batched()
	if err != nil {
		return err
	}

	var innerErr error
	if err := source.Iterate(EmptyPrefix, func(key, value Value) bool {
		currentBatchSize++

		if err := batchedMutation.Set(key, value); err != nil {
			innerErr = err
		}

		if batchedSize != 0 && currentBatchSize >= batchedSize {
			if err := batchedMutation.Commit(); err != nil {
				innerErr = err
			}

			// init new batched mutation
			currentBatchSize = 0
			batchedMutation, err = target.Batched()
			if err != nil {
				innerErr = err
			}
		}

		return innerErr == nil
	}); err != nil {
		batchedMutation.Cancel()

		return err
	}

	if innerErr != nil {
		batchedMutation.Cancel()

		return innerErr
	}

	if err := batchedMutation.Commit(); err != nil {
		return err
	}

	return target.Flush()
}

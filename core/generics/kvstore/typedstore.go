package kvstore

import (
	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/core/generics/constraints"
	"github.com/iotaledger/hive.go/core/kvstore"
)

// TypedStore is a generically typed wrapper around a KVStore that abstracts serialization away.
type TypedStore[K, V any, KPtr constraints.Serializable[K], VPtr constraints.Serializable[V]] struct {
	kv kvstore.KVStore
}

// NewTypedStore is the constructor for TypedStore.
func NewTypedStore[K, V any, KPtr constraints.Serializable[K], VPtr constraints.Serializable[V]](kv kvstore.KVStore) *TypedStore[K, V, KPtr, VPtr] {
	return &TypedStore[K, V, KPtr, VPtr]{
		kv: kv,
	}
}

// Get gets the given key or an error if an error occurred.
func (t *TypedStore[K, V, KPtr, VPtr]) Get(key K) (value VPtr, err error) {
	keyBytes, err := (KPtr)(&key).Bytes()
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode key")
	}

	valueBytes, err := t.kv.Get(keyBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve from KV store")
	}

	value = new(V)
	if _, err = value.FromBytes(valueBytes); err != nil {
		return nil, errors.Wrap(err, "failed to decode value")
	}

	return value, nil
}

// Set sets the given key and value.
func (t *TypedStore[K, V, KPtr, VPtr]) Set(key K, value VPtr) (err error) {
	keyBytes, err := (KPtr)(&key).Bytes()
	if err != nil {
		return errors.Wrap(err, "failed to encode key")
	}

	valueBytes, err := value.Bytes()
	if err != nil {
		return errors.Wrap(err, "failed to encode value")
	}

	err = t.kv.Set(keyBytes, valueBytes)
	if err != nil {
		return errors.Wrap(err, "failed to store in KV store")
	}

	return nil
}

// Delete deletes the given key from the store:wq.
func (t *TypedStore[K, V, KPtr, VPtr]) Delete(key K) (err error) {
	keyBytes, err := (KPtr)(&key).Bytes()
	if err != nil {
		return errors.Wrap(err, "failed to encode key")
	}

	err = t.kv.Delete(keyBytes)
	if err != nil {
		return errors.Wrap(err, "failed to delete entry from KV store")
	}

	return nil
}

func (t *TypedStore[K, V, KPtr, VPtr]) Iterate(prefix kvstore.KeyPrefix, callback func(key K, value VPtr) (advance bool), direction ...kvstore.IterDirection) (err error) {
	if iterationErr := t.kv.Iterate(prefix, func(key kvstore.Key, value kvstore.Value) bool {
		var keyDecoded KPtr = new(K)
		if _, err = keyDecoded.FromBytes(key); err != nil {
			return false
		}

		var valueDecoded VPtr = new(V)
		if _, err = valueDecoded.FromBytes(value); err != nil {
			return false
		}

		return callback(*keyDecoded, valueDecoded)
	}, direction...); iterationErr != nil {
		return errors.Wrap(iterationErr, "failed to iterate over KV store")
	}

	return
}

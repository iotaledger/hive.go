package kvstore

import (
	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/core/kvstore"
)

// TypedStore is a generically typed wrapper around a KVStore that abstracts serialization away.
type TypedStore[K KeyType, V any, VPtr ValuePtrType[V]] struct {
	kv kvstore.KVStore
}

// NewTypedStore is the constructor for TypedStore.
func NewTypedStore[K KeyType, V any, VPtr ValuePtrType[V]](kv kvstore.KVStore) *TypedStore[K, V, VPtr] {
	return &TypedStore[K, V, VPtr]{
		kv: kv,
	}
}

// Get gets the given key or an error if an error occurred.
func (t *TypedStore[K, V, VPtr]) Get(key K) (value VPtr, err error) {
	keyBytes, err := key.Bytes()
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
func (t *TypedStore[K, V, VPtr]) Set(key K, value VPtr) (err error) {
	keyBytes, err := key.Bytes()
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

// KeyType is a type constraints for the keys of the TypedStore.
type KeyType interface {
	Bytes() ([]byte, error)
}

// ValuePtrType is a type constraints for values of the TypedStore.
type ValuePtrType[V any] interface {
	*V

	FromBytes([]byte) (consumedBytes int, err error)
	Bytes() ([]byte, error)
}

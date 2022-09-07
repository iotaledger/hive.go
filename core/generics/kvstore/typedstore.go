package kvstore

import (
	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/core/kvstore"
)

// TypedStore is a generically typed wrapper around a KVStore that abstracts serialization away.
type TypedStore[K serializable, V deserializable] struct {
	kv kvstore.KVStore
}

// NewTypedStore is the constructor for TypedStore.
func NewTypedStore[K serializable, V deserializable](kv kvstore.KVStore) *TypedStore[K, V] {
	return &TypedStore[K, V]{
		kv: kv,
	}
}

// Get gets the given key or an error if an error occurred.
func (t *TypedStore[K, V]) Get(key K) (value V, err error) {
	keyBytes, err := key.Bytes()
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode key")
	}

	valueBytes, err := t.kv.Get(keyBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve from KV store")
	}

	_, err = value.FromBytes(valueBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode value")
	}

	return value, nil
}

// Set sets the given key and value.
func (t *TypedStore[K, V]) Set(key K, value V) (err error) {
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

// deserializable is an interface that for a type that is serializable and deserializable.
type deserializable interface {
	FromBytes(b []byte) (consumedBytes int, err error)

	serializable
}

// serializable is an interface for a type that is serializable.
type serializable interface {
	Bytes() ([]byte, error)
}

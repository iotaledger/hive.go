package kvstore

import (
	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/core/kvstore"
)

// DeAndSerializable is an interface for a type that is Serializable and Deserializable.
type DeAndSerializable[A any] interface {
	DeSerializable[A]
	Serializable
}

// DeSerializable is an interface for a type that is Serializable and Deserializable.
type DeSerializable[A any] interface {
	*A

	FromBytes(b []byte) error
}

// Serializable is an interface for a type that is Serializable and Deserializable.
type Serializable interface {
	Bytes() ([]byte, error)
}

// TypedStore is a generically typed wrapper around a kvstore.KVStore that abstracts serialization away.
type TypedStore[K Serializable, V DeAndSerializable[V]] struct {
	kv kvstore.KVStore
}

// NewTypedStore is the constructor for TypedStore.
func NewTypedStore[K Serializable, V DeAndSerializable[V]](kv kvstore.KVStore) *TypedStore[K, V] {
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

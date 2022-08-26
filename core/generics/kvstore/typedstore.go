package kvstore

import (
	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/core/serix"
)

// DeAndSerializable is an interface for a type that is Serializable and Deserializable.
type DeAndSerializable interface {
	serix.Serializable
	serix.Deserializable
}

// TypedStore is a generically typed wrapper around a kvstore.KVStore that abstracts serialization away.
type TypedStore[K, V DeAndSerializable] struct {
	kv kvstore.KVStore
}

// NewTypedStore is the constructor for TypedStore.
func NewTypedStore[K, V DeAndSerializable](kv kvstore.KVStore) *TypedStore[K, V] {
	return &TypedStore[K, V]{
		kv: kv,
	}
}

// Get gets the given key or an error if an error occurred.
func (t *TypedStore[K, V]) Get(key K) (value V, exists bool, err error) {
	keyBytes, err := key.Encode()
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to encode key")
	}

	valueBytes, err := t.kv.Get(keyBytes)
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to retrieve from KV store")
	}
	if valueBytes == nil {
		return nil, false, nil
	}

	_, err = value.Decode(valueBytes)
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to decode value")
	}

	return value, true, nil
}

// Set sets the given key and value.
func (t *TypedStore[K, V]) Set(key K, value V) (err error) {
	keyBytes, err := key.Encode()
	if err != nil {
		return errors.Wrap(err, "failed to encode key")
	}

	valueBytes, err := value.Encode()
	if err != nil {
		return errors.Wrap(err, "failed to encode value")
	}

	err = t.kv.Set(keyBytes, valueBytes)
	if err != nil {
		return errors.Wrap(err, "failed to store in KV store")
	}

	return nil
}

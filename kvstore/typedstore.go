package kvstore

import (
	"github.com/iotaledger/hive.go/ierrors"
)

type (
	ObjectToBytes[O any] func(O) ([]byte, error)
	BytesToObject[O any] func([]byte) (object O, consumed int, err error)
)

// TypedStore is a generically typed wrapper around a KVStore that abstracts serialization away.
type TypedStore[K, V any] struct {
	kv KVStore

	keyToBytes   ObjectToBytes[K]
	bytesToKey   BytesToObject[K]
	valueToBytes ObjectToBytes[V]
	bytesToValue BytesToObject[V]
}

// NewTypedStore is the constructor for TypedStore.
func NewTypedStore[K, V any](
	kv KVStore,
	keyToBytes ObjectToBytes[K],
	bytesToKey BytesToObject[K],
	valueToBytes ObjectToBytes[V],
	bytesToValue BytesToObject[V],
) *TypedStore[K, V] {
	return &TypedStore[K, V]{
		kv:           kv,
		keyToBytes:   keyToBytes,
		bytesToKey:   bytesToKey,
		valueToBytes: valueToBytes,
		bytesToValue: bytesToValue,
	}
}

func (t *TypedStore[K, V]) KVStore() KVStore {
	return t.kv
}

// Get gets the given key or an error if an error occurred.
func (t *TypedStore[K, V]) Get(key K) (value V, err error) {
	keyBytes, err := t.keyToBytes(key)
	if err != nil {
		return value, ierrors.Wrap(err, "failed to encode key")
	}

	valueBytes, err := t.kv.Get(keyBytes)
	if err != nil {
		return value, ierrors.Wrap(err, "failed to retrieve from KV store")
	}

	v, _, err := t.bytesToValue(valueBytes)
	if err != nil {
		return value, ierrors.Wrap(err, "failed to decode value")
	}

	return v, nil
}

// Has checks whether the given key exists.
func (t *TypedStore[K, V]) Has(key K) (has bool, err error) {
	keyBytes, err := t.keyToBytes(key)
	if err != nil {
		return false, ierrors.Wrap(err, "failed to encode key")
	}

	return t.kv.Has(keyBytes)
}

// Set sets the given key and value.
func (t *TypedStore[K, V]) Set(key K, value V) (err error) {
	keyBytes, err := t.keyToBytes(key)
	if err != nil {
		return ierrors.Wrap(err, "failed to encode key")
	}

	valueBytes, err := t.valueToBytes(value)
	if err != nil {
		return ierrors.Wrap(err, "failed to encode value")
	}

	err = t.kv.Set(keyBytes, valueBytes)
	if err != nil {
		return ierrors.Wrap(err, "failed to store in KV store")
	}

	return nil
}

// Delete deletes the given key from the store.
func (t *TypedStore[K, V]) Delete(key K) (err error) {
	keyBytes, err := t.keyToBytes(key)
	if err != nil {
		return ierrors.Wrap(err, "failed to encode key")
	}

	err = t.kv.Delete(keyBytes)
	if err != nil {
		return ierrors.Wrap(err, "failed to delete entry from KV store")
	}

	return nil
}

func (t *TypedStore[K, V]) Iterate(prefix KeyPrefix, callback func(key K, value V) (advance bool), direction ...IterDirection) (err error) {
	var innerErr error
	if iterationErr := t.kv.Iterate(prefix, func(key Key, value Value) bool {
		keyDecoded, _, keyErr := t.bytesToKey(key)
		if keyErr != nil {
			innerErr = keyErr

			return false
		}

		valueDecoded, _, valueErr := t.bytesToValue(value)
		if valueErr != nil {
			innerErr = valueErr

			return false
		}

		return callback(keyDecoded, valueDecoded)
	}, direction...); iterationErr != nil {
		return ierrors.Wrap(iterationErr, "failed to iterate over KV store")
	}

	return innerErr
}

func (t *TypedStore[K, V]) IterateKeys(prefix KeyPrefix, callback func(key K) (advance bool), direction ...IterDirection) (err error) {
	var innerErr error
	if iterationErr := t.kv.IterateKeys(prefix, func(key Key) bool {
		keyDecoded, _, keyErr := t.bytesToKey(key)
		if keyErr != nil {
			innerErr = keyErr

			return false
		}

		return callback(keyDecoded)
	}, direction...); iterationErr != nil {
		return ierrors.Wrap(iterationErr, "failed to iterate keys over KV store")
	}

	return innerErr
}

func (t *TypedStore[K, V]) DeletePrefix(prefix KeyPrefix) error {
	return t.kv.DeletePrefix(prefix)
}

func (t *TypedStore[K, V]) Clear() error {
	return t.kv.Clear()
}

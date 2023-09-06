package kvstore

import "github.com/iotaledger/hive.go/ierrors"

// TypedValue is a generically typed wrapper around a KVStore that provides access to a single value.
type TypedValue[V any] struct {
	kv       KVStore
	keyBytes []byte

	vToBytes ObjectToBytes[V]
	bytesToV BytesToObject[V]
}

// NewTypedValue is the constructor for TypedValue.
func NewTypedValue[V any](
	kv KVStore,
	keyBytes []byte,
	vToBytes ObjectToBytes[V],
	bytesToV BytesToObject[V],
) *TypedValue[V] {
	return &TypedValue[V]{
		kv:       kv,
		keyBytes: keyBytes,
		vToBytes: vToBytes,
		bytesToV: bytesToV,
	}
}

func (t *TypedValue[V]) KVStore() KVStore {
	return t.kv
}

// Get gets the given key or an error if an error occurred.
func (t *TypedValue[V]) Get() (value V, err error) {
	valueBytes, err := t.kv.Get(t.keyBytes)
	if err != nil {
		return value, ierrors.Wrap(err, "failed to retrieve from KV store")
	}

	v, _, err := t.bytesToV(valueBytes)
	if err != nil {
		return value, ierrors.Wrap(err, "failed to decode value")
	}

	return v, nil
}

// Has checks whether the given key exists.
func (t *TypedValue[V]) Has() (has bool, err error) {
	return t.kv.Has(t.keyBytes)
}

// Set sets the given key and value.
func (t *TypedValue[V]) Set(value V) (err error) {
	valueBytes, err := t.vToBytes(value)
	if err != nil {
		return ierrors.Wrap(err, "failed to encode value")
	}

	err = t.kv.Set(t.keyBytes, valueBytes)
	if err != nil {
		return ierrors.Wrap(err, "failed to store in KV store")
	}

	return nil
}

// Delete deletes the given key from the store.
func (t *TypedValue[V]) Delete() (err error) {
	err = t.kv.Delete(t.keyBytes)
	if err != nil {
		return ierrors.Wrap(err, "failed to delete entry from KV store")
	}

	return nil
}

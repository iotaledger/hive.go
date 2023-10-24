package kvstore

import (
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/runtime/syncutils"
)

// TypedValue is a generically typed wrapper around a KVStore that provides access to a single value.
type TypedValue[V any] struct {
	kv       KVStore
	keyBytes []byte

	vToBytes ObjectToBytes[V]
	bytesToV BytesToObject[V]

	valueCached       *V
	valueExistsCached *bool
	mutex             syncutils.Mutex
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

// KVStore returns the underlying KVStore.
func (t *TypedValue[V]) KVStore() KVStore {
	return t.kv
}

// Get gets the given key or an error if an error occurred.
func (t *TypedValue[V]) Get() (value V, err error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.valueExistsCached != nil && !*t.valueExistsCached {
		return value, ErrKeyNotFound
	}

	if t.valueCached != nil {
		return *t.valueCached, nil
	}

	if valueBytes, valueBytesErr := t.kv.Get(t.keyBytes); valueBytesErr != nil {
		if ierrors.Is(valueBytesErr, ErrKeyNotFound) {
			t.valueExistsCached = &falsePtr
		}

		return value, ierrors.Wrap(valueBytesErr, "failed to retrieve value from KV store")
	} else if value, _, err = t.bytesToV(valueBytes); err != nil {
		return value, ierrors.Wrap(err, "failed to decode value")
	}

	t.valueCached = &value
	t.valueExistsCached = &truePtr

	return value, nil
}

// Has checks whether the given key exists.
func (t *TypedValue[V]) Has() (has bool, err error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.valueExistsCached != nil {
		return *t.valueExistsCached, nil
	} else if has, err = t.kv.Has(t.keyBytes); err != nil {
		return false, ierrors.Wrap(err, "failed to check whether key exists")
	}

	t.valueExistsCached = &has

	return has, nil
}

// Compute atomically computes and sets a new value based on the current value and some provided computation function.
func (t *TypedValue[V]) Compute(computeFunc func(currentValue V, exists bool) (newValue V, err error)) (newValue V, err error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	var currentValue V
	var exists bool

	if t.valueCached != nil {
		currentValue = *t.valueCached
		exists = true
	} else if t.valueExistsCached == nil || *t.valueExistsCached {
		if valueBytes, valueBytesErr := t.kv.Get(t.keyBytes); valueBytesErr != nil {
			if !ierrors.Is(valueBytesErr, ErrKeyNotFound) {
				return newValue, ierrors.Wrap(valueBytesErr, "failed to retrieve value from KV store")
			}
		} else if currentValue, _, err = t.bytesToV(valueBytes); err != nil {
			return newValue, ierrors.Wrap(err, "failed to decode value")
		} else {
			exists = true
		}
	}

	if newValue, err = computeFunc(currentValue, exists); err != nil {
		return newValue, ierrors.Wrap(err, "failed to compute new value")
	}

	t.valueCached = &newValue
	t.valueExistsCached = &truePtr

	return newValue, nil
}

// Set sets the given key and value.
func (t *TypedValue[V]) Set(value V) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if valueBytes, err := t.vToBytes(value); err != nil {
		return ierrors.Wrap(err, "failed to encode value")
	} else if err = t.kv.Set(t.keyBytes, valueBytes); err != nil {
		return ierrors.Wrap(err, "failed to store in KV store")
	}

	t.valueCached = &value
	t.valueExistsCached = &truePtr

	return nil
}

// Delete deletes the given key from the store.
func (t *TypedValue[V]) Delete() (err error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if err = t.kv.Delete(t.keyBytes); err != nil {
		return ierrors.Wrap(err, "failed to delete entry from KV store")
	}

	t.valueCached = nil
	t.valueExistsCached = &falsePtr

	return nil
}

// truePtr is a pointer to a true value.
var truePtr = true

// falsePtr is a pointer to a false value.
var falsePtr = false

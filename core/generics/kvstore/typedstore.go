package kvstore

import (
	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/core/serix"
)

type DeAndSerializable interface {
	serix.Serializable
	serix.Deserializable
}

type TypedStored[K, V DeAndSerializable] struct {
	kv kvstore.KVStore
}

func (t *TypedStored[K, V]) Get(key K) (value V, exists bool, err error) {
	keyBytes, err := key.Encode()
	if err != nil {
		return nil, false, err
	}

	valueBytes, err := t.kv.Get(keyBytes)
	if err != nil {
		return nil, false, err
	}
	if valueBytes == nil {
		return nil, false, nil
	}

	_, err = value.Decode(valueBytes)
	if err != nil {
		return nil, false, err
	}
	return value, true, nil
}

func (t *TypedStored[K, V]) Set(key K, value V) (err error) {
	keyBytes, err := key.Encode()
	if err != nil {
		return err
	}
	valueBytes, err := value.Encode()
	if err != nil {
		return err
	}
	err = t.kv.Set(keyBytes, valueBytes)
	return err
}

package model

import (
	"context"
	"sync"

	"github.com/cockroachdb/errors"

	"github.com/iotaledger/hive.go/generics/objectstorage"
	"github.com/iotaledger/hive.go/serix"
)

// region Model ////////////////////////////////////////////////////////////////////////////////////////////////////////

type Model[Key any, T any] struct {
	id      *Key
	idFunc  func(model *T) Key
	idMutex sync.RWMutex
	m       T

	sync.RWMutex
	objectstorage.StorableObjectFlags
}

func NewModel[Key any, T any](model T, optionalIDFunc ...func(model *T) Key) Model[Key, T] {
	if len(optionalIDFunc) == 0 {
		return Model[Key, T]{
			m: model,
			idFunc: func(model *T) (key Key) {
				return key
			},
		}
	}

	return Model[Key, T]{
		m:      model,
		idFunc: optionalIDFunc[0],
	}
}

func (m *Model[Key, T]) ID() (id Key) {
	m.idMutex.RLock()
	if m.id != nil {
		defer m.idMutex.RUnlock()
		return *m.id
	}
	m.idMutex.RUnlock()

	m.idMutex.Lock()
	defer m.idMutex.Unlock()
	if m.id != nil {
		return *m.id
	}

	id = m.idFunc(&m.m)
	m.id = &id

	return id
}

func (m *Model[Key, T]) SetID(id Key) {
	m.idMutex.Lock()
	defer m.idMutex.Unlock()

	m.id = &id
}

func (m *Model[Key, T]) Decode(b []byte) (int, error) {
	m.Lock()
	defer m.Unlock()

	return serix.DefaultAPI.Decode(context.Background(), b, &m.m)
}

func (m *Model[Key, T]) Encode() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()

	return serix.DefaultAPI.Encode(context.Background(), &m.m, serix.WithValidation())
}

func (m *Model[Key, T]) FromObjectStorage(key, data []byte) (err error) {
	m.idMutex.Lock()
	defer m.idMutex.Unlock()
	if _, err = serix.DefaultAPI.Decode(context.Background(), key, &m.id); err != nil {
		return errors.Errorf("failed to decode key: %w", err)
	}

	m.Lock()
	defer m.Unlock()
	if _, err = serix.DefaultAPI.Decode(context.Background(), data, &m.m); err != nil {
		return errors.Errorf("failed to decode m: %w", err)
	}

	return nil
}

func (m *Model[Key, T]) ObjectStorageKey() (key []byte) {
	key, err := serix.DefaultAPI.Encode(context.Background(), m.ID(), serix.WithValidation())
	if err != nil {
		panic(err)
	}

	return key
}

func (m *Model[Key, T]) ObjectStorageValue() (value []byte) {
	value, err := m.Encode()
	if err != nil {
		panic(err)
	}

	return value
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

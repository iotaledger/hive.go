package model

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/cockroachdb/errors"

	"github.com/iotaledger/hive.go/generics/objectstorage"
	"github.com/iotaledger/hive.go/serix"
)

type Model[IDType any, ModelType any] struct {
	id      *IDType
	idFunc  func(model *ModelType) IDType
	idMutex sync.RWMutex
	M       ModelType

	sync.RWMutex
	objectstorage.StorableObjectFlags
}

func NewModel[IDType any, ModelType any](model ModelType, optionalIDFunc ...func(model *ModelType) IDType) Model[IDType, ModelType] {
	if len(optionalIDFunc) == 0 {
		optionalIDFunc = append(optionalIDFunc, func(model *ModelType) (key IDType) {
			return key
		})
	}

	return Model[IDType, ModelType]{
		M:      model,
		idFunc: optionalIDFunc[0],
	}
}

func (m *Model[IDType, ModelType]) ID() (id IDType) {
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

	id = m.idFunc(&m.M)
	m.id = &id

	return id
}

func (m *Model[IDType, ModelType]) SetID(id IDType) {
	m.idMutex.Lock()
	defer m.idMutex.Unlock()

	m.id = &id
}

func (m *Model[IDType, ModelType]) Decode(b []byte) (int, error) {
	m.Lock()
	defer m.Unlock()

	return serix.DefaultAPI.Decode(context.Background(), b, &m.M)
}

func (m *Model[IDType, ModelType]) Encode() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()

	return serix.DefaultAPI.Encode(context.Background(), &m.M, serix.WithValidation())
}

func (m *Model[IDType, ModelType]) FromObjectStorage(key, data []byte) (err error) {
	m.idMutex.Lock()
	defer m.idMutex.Unlock()
	if _, err = serix.DefaultAPI.Decode(context.Background(), key, &m.id); err != nil {
		return errors.Errorf("failed to decode key: %w", err)
	}

	m.Lock()
	defer m.Unlock()
	if _, err = serix.DefaultAPI.Decode(context.Background(), data, &m.M); err != nil {
		return errors.Errorf("failed to decode m: %w", err)
	}

	return nil
}

func (m *Model[IDType, ModelType]) ObjectStorageKey() (key []byte) {
	key, err := serix.DefaultAPI.Encode(context.Background(), m.ID(), serix.WithValidation())
	if err != nil {
		panic(err)
	}

	return key
}

func (m *Model[IDType, ModelType]) ObjectStorageValue() (value []byte) {
	value, err := m.Encode()
	if err != nil {
		panic(err)
	}

	return value
}

func (m *Model[IDType, ModelType]) String() string {
	return fmt.Sprintf("Model[%s] {\n\tid: %+v\n\tmodel: %+v\n}", reflect.TypeOf(m.M).Name(), m.id, m.M)
}

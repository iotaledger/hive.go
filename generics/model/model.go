package model

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/cockroachdb/errors"

	"github.com/iotaledger/hive.go/byteutils"
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

func NewModel[IDType any, ModelType any](model ModelType, optionalIDFunc ...func(model *ModelType) IDType) (new Model[IDType, ModelType]) {
	if len(optionalIDFunc) == 0 {
		optionalIDFunc = append(optionalIDFunc, func(model *ModelType) (key IDType) {
			return key
		})
	}

	new = Model[IDType, ModelType]{
		M:      model,
		idFunc: optionalIDFunc[0],
	}
	new.SetModified()
	new.Persist()

	return new
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

// region KeyModel /////////////////////////////////////////////////////////////////////////////////////////////////////

type ReferenceModel[SourceIDType, TargetIDType any] struct {
	SourceID SourceIDType
	TargetID TargetIDType

	sync.RWMutex
	objectstorage.StorableObjectFlags
}

func NewReferenceModel[SourceIDType, TargetIDType any](s SourceIDType, t TargetIDType) (new ReferenceModel[SourceIDType, TargetIDType]) {
	new = ReferenceModel[SourceIDType, TargetIDType]{
		SourceID: s,
		TargetID: t,
	}
	new.SetModified()
	new.Persist()

	return new
}

func (m *ReferenceModel[SourceIDType, TargetIDType]) Decode(b []byte) (int, error) {
	m.Lock()
	defer m.Unlock()

	consumedBytesSource, err := serix.DefaultAPI.Decode(context.Background(), b, &m.SourceID)
	if err != nil {
		return consumedBytesSource, errors.Errorf("failed to decode SourceID: %w", err)
	}

	consumedBytesTarget, err := serix.DefaultAPI.Decode(context.Background(), b[consumedBytesSource:], &m.TargetID)
	if err != nil {
		return consumedBytesTarget, errors.Errorf("failed to decode TargetID: %w", err)
	}

	return consumedBytesSource + consumedBytesTarget, nil
}

func (m *ReferenceModel[SourceIDType, TargetIDType]) Encode() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()

	sourceBytes, err := serix.DefaultAPI.Encode(context.Background(), &m.SourceID, serix.WithValidation())
	if err != nil {
		return nil, errors.Errorf("failed to encode source: %w", err)
	}
	targetBytes, err := serix.DefaultAPI.Encode(context.Background(), &m.TargetID, serix.WithValidation())
	if err != nil {
		return nil, errors.Errorf("failed to encode target: %w", err)
	}

	return byteutils.ConcatBytes(sourceBytes, targetBytes), nil
}

func (m *ReferenceModel[SourceIDType, TargetIDType]) FromObjectStorage(key, _ []byte) (err error) {
	_, err = m.Decode(key)
	return err
}

func (m *ReferenceModel[SourceIDType, TargetIDType]) ObjectStorageKey() (key []byte) {
	key, err := m.Encode()
	if err != nil {
		panic(err)
	}

	return key
}

func (m *ReferenceModel[SourceIDType, TargetIDType]) ObjectStorageValue() (value []byte) {
	return nil
}

func (m *ReferenceModel[SourceIDType, TargetIDType]) String() string {
	return fmt.Sprintf("ReferenceModel[%s,%s] {\n\tSourceID: %+v\n\tTargetID: %+v\n}",
		reflect.TypeOf(m.SourceID).Name(), reflect.TypeOf(m.TargetID).Name(), m.SourceID, m.TargetID)
}

func (m ReferenceModel[SourceIDType, TargetIDType]) PartitionKey() []int {
	var s SourceIDType
	sourceBytes, err := serix.DefaultAPI.Encode(context.Background(), s, serix.WithValidation())
	if err != nil {
		panic(err)
	}

	var t TargetIDType
	targetBytes, err := serix.DefaultAPI.Encode(context.Background(), t, serix.WithValidation())
	if err != nil {
		panic(err)
	}
	return []int{len(sourceBytes), len(targetBytes)}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

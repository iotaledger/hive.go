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

type Model[IDType, ModelType any] struct {
	id      *IDType
	idFunc  func(model *ModelType) IDType
	idMutex sync.RWMutex
	M       ModelType `serix:"0"`

	sync.RWMutex
	objectstorage.StorableObjectFlags
}

func NewModel[IDType, ModelType any](model ModelType, optionalIDFunc ...func(model *ModelType) IDType) (new Model[IDType, ModelType]) {
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
	m.RLock()
	defer m.RUnlock()

	value, err := serix.DefaultAPI.Encode(context.Background(), &m.M, serix.WithValidation())
	if err != nil {
		panic(err)
	}

	return value
}

func (m *Model[IDType, ModelType]) Bytes() (bytes []byte, err error) {
	m.RLock()
	defer m.RUnlock()

	return serix.DefaultAPI.Encode(context.Background(), m.M, serix.WithValidation())
}

func (m *Model[IDType, ModelType]) String() string {
	return fmt.Sprintf("Model[%s] {\n\tid: %+v\n\tmodel: %+v\n}", reflect.TypeOf(m.M).Name(), m.id, m.M)
}

// region ReferenceModel ///////////////////////////////////////////////////////////////////////////////////////////////

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

func (m *ReferenceModel[SourceIDType, TargetIDType]) FromObjectStorage(key, _ []byte) (err error) {
	m.Lock()
	defer m.Unlock()

	consumedBytesSource, err := serix.DefaultAPI.Decode(context.Background(), key, &m.SourceID, serix.WithValidation())
	if err != nil {
		return errors.Errorf("failed to decode SourceID: %w", err)
	}

	_, err = serix.DefaultAPI.Decode(context.Background(), key[consumedBytesSource:], &m.TargetID, serix.WithValidation())
	if err != nil {
		return errors.Errorf("failed to decode TargetID: %w", err)
	}

	return err
}

func (m *ReferenceModel[SourceIDType, TargetIDType]) ObjectStorageKey() (key []byte) {
	m.RLock()
	defer m.RUnlock()

	sourceBytes, err := serix.DefaultAPI.Encode(context.Background(), &m.SourceID, serix.WithValidation())
	if err != nil {
		panic(errors.Errorf("failed to encode source: %w", err))
	}
	targetBytes, err := serix.DefaultAPI.Encode(context.Background(), &m.TargetID, serix.WithValidation())
	if err != nil {
		panic(errors.Errorf("failed to encode target: %w", err))
	}

	return byteutils.ConcatBytes(sourceBytes, targetBytes)
}

func (m *ReferenceModel[SourceIDType, TargetIDType]) ObjectStorageValue() (value []byte) {
	return nil
}

func (m ReferenceModel[SourceIDType, TargetIDType]) KeyPartitions() []int {
	var s SourceIDType
	var t TargetIDType

	return []int{len(serix.Encode(s)), len(serix.Encode(t))}
}

func (m *ReferenceModel[SourceIDType, TargetIDType]) String() string {
	return fmt.Sprintf("ReferenceModel[%s,%s] {\n\tSourceID: %+v\n\tTargetID: %+v\n}",
		reflect.TypeOf(m.SourceID).Name(), reflect.TypeOf(m.TargetID).Name(), m.SourceID, m.TargetID)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ReferenceModelWithMetadata ///////////////////////////////////////////////////////////////////////////////////

type ReferenceModelWithMetadata[SourceIDType, TargetIDType, ModelType any] struct {
	SourceID SourceIDType
	TargetID TargetIDType
	M        ModelType `serix:"0"`

	sync.RWMutex
	objectstorage.StorableObjectFlags
}

func NewReferenceModelWithMetadata[SourceIDType, TargetIDType, ModelType any](s SourceIDType, t TargetIDType, m ModelType) (new ReferenceModelWithMetadata[SourceIDType, TargetIDType, ModelType]) {
	new = ReferenceModelWithMetadata[SourceIDType, TargetIDType, ModelType]{
		SourceID: s,
		TargetID: t,
		M:        m,
	}
	new.SetModified()
	new.Persist()

	return new
}

func (m *ReferenceModelWithMetadata[SourceIDType, TargetIDType, ModelType]) FromObjectStorage(key, data []byte) (err error) {
	m.Lock()
	defer m.Unlock()

	consumedSourceIDBytes, err := serix.DefaultAPI.Decode(context.Background(), key, &m.SourceID, serix.WithValidation())
	if err != nil {
		return errors.Errorf("failed to decode source ID: %w", err)
	}
	_, err = serix.DefaultAPI.Decode(context.Background(), key[consumedSourceIDBytes:], &m.TargetID, serix.WithValidation())
	if err != nil {
		return errors.Errorf("failed to decode target ID: %w", err)
	}

	if _, err = serix.DefaultAPI.Decode(context.Background(), data, &m.M, serix.WithValidation()); err != nil {
		return errors.Errorf("failed to decode model: %w", err)
	}

	return nil
}

func (m *ReferenceModelWithMetadata[SourceIDType, TargetIDType, ModelType]) ObjectStorageKey() (key []byte) {
	m.RLock()
	defer m.RUnlock()

	return byteutils.ConcatBytes(serix.Encode(m.SourceID), serix.Encode(m.TargetID))
}

func (m *ReferenceModelWithMetadata[SourceIDType, TargetIDType, ModelType]) ObjectStorageValue() (value []byte) {
	m.RLock()
	defer m.RUnlock()

	return serix.Encode(&m.M)
}

func (m ReferenceModelWithMetadata[SourceIDType, TargetIDType, ModelType]) KeyPartitions() []int {
	var s SourceIDType
	var t TargetIDType

	return []int{len(serix.Encode(s)), len(serix.Encode(t))}
}

func (m *ReferenceModelWithMetadata[SourceIDType, TargetIDType, ModelType]) Bytes() (bytes []byte, err error) {
	m.RLock()
	defer m.RUnlock()

	return serix.DefaultAPI.Encode(context.Background(), m.M, serix.WithValidation())
}

func (m *ReferenceModelWithMetadata[SourceIDType, TargetIDType, ModelType]) String() string {
	m.RLock()
	defer m.RUnlock()

	return fmt.Sprintf("ReferenceModelWithMetadata[%s, %s, %s] {\n\tsourceID: %+v\n\ttargetID: %+v\n\tmodel: %+v\n}", reflect.TypeOf(m.SourceID).Name(), reflect.TypeOf(m.TargetID).Name(), reflect.TypeOf(m.M).Name(), m.SourceID, m.TargetID, m.M)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

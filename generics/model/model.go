package model

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/iotaledger/hive.go/serix"
)

// Model is the base type for all models. It should be embedded in a wrapper type.
// It provides locking and serialization primitives.
type Model[OuterModelType any, OuterModelPtrType outerModelPtr[OuterModelType, InnerModelType], InnerModelType any] struct {
	M InnerModelType `serix:"0"`

	mutex sync.RWMutex
}

// New returns a new model instance.
func New[OuterModelType, InnerModelType any, OuterModelPtrType outerModelPtr[OuterModelType, InnerModelType]](model *InnerModelType) (newInstance *OuterModelType) {
	newInstance = new(OuterModelType)
	(OuterModelPtrType)(newInstance).setM(model)

	return newInstance
}

// FromBytes deserializes a model from a byte slice.
func (m *Model[OuterModelType, OuterModelPtrType, InnerModelType]) FromBytes(bytes []byte) (err error) {
	m.Lock()
	defer m.Unlock()

	outerModel := new(OuterModelType)
	_, err = serix.DefaultAPI.Decode(context.Background(), bytes, outerModel, serix.WithValidation())
	m.M = *(OuterModelPtrType)(outerModel).m()

	return
}

// RLock read-locks the Model.
func (m *Model[OuterModelType, OuterModelPtrType, InnerModelType]) RLock() {
	m.mutex.RLock()
}

// RUnlock read-unlocks the Model.
func (m *Model[OuterModelType, OuterModelPtrType, InnerModelType]) RUnlock() {
	m.mutex.RUnlock()
}

// Lock write-locks the Model.
func (m *Model[OuterModelType, OuterModelPtrType, InnerModelType]) Lock() {
	m.mutex.Lock()
}

// Unlock write-unlocks the Model.
func (m *Model[OuterModelType, OuterModelPtrType, InnerModelType]) Unlock() {
	m.mutex.Unlock()
}

// Bytes serializes a model to a byte slice.
func (m *Model[OuterModelType, OuterModelPtrType, InnerModelType]) Bytes() (bytes []byte, err error) {
	m.RLock()
	defer m.RUnlock()

	outerModel := new(OuterModelType)
	(OuterModelPtrType)(outerModel).setM(&m.M)

	return serix.DefaultAPI.Encode(context.Background(), outerModel, serix.WithValidation())
}

// String returns a string representation of the model.
func (m *Model[OuterModelType, OuterModelPtrType, InnerModelType]) String() string {
	var outerModelType OuterModelType
	return fmt.Sprintf("Model[%s] %+v", reflect.TypeOf(outerModelType).Name(), m.M)
}

func (m *Model[OuterModelType, OuterModelPtrType, InnerModelType]) setM(innerModel *InnerModelType) {
	m.M = *innerModel
}

func (m *Model[OuterModelType, OuterModelPtrType, InnerModelType]) m() *InnerModelType {
	return &m.M
}

type outerModelPtr[OuterModelType any, InnerModelType any] interface {
	*OuterModelType

	setM(*InnerModelType)
	m() *InnerModelType
}

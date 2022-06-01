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
type Model[ModelType any] struct {
	M ModelType `serix:"0"`

	mutex sync.RWMutex
}

// New creates a new model instance.
func New[ModelType any](model ModelType) (newInstance Model[ModelType]) {
	newInstance = Model[ModelType]{
		M: model,
	}

	return newInstance
}

// FromBytes deserializes a model from a byte slice.
func (m *Model[ModelType]) FromBytes(bytes []byte) (err error) {
	m.Lock()
	defer m.Unlock()

	_, err = serix.DefaultAPI.Decode(context.Background(), bytes, &m.M, serix.WithValidation())
	return
}

// RLock read-locks the Model.
func (m *Model[ModelType]) RLock() {
	m.mutex.RLock()
}

// RUnlock read-unlocks the Model.
func (m *Model[ModelType]) RUnlock() {
	m.mutex.RUnlock()
}

// Lock write-locks the Model.
func (m *Model[ModelType]) Lock() {
	m.mutex.Lock()
}

// Unlock write-unlocks the Model.
func (m *Model[ModelType]) Unlock() {
	m.mutex.Unlock()
}

// Bytes serializes a model to a byte slice.
func (m *Model[ModelType]) Bytes() (bytes []byte, err error) {
	m.RLock()
	defer m.RUnlock()

	return serix.DefaultAPI.Encode(context.Background(), m.M, serix.WithValidation())
}

// String returns a string representation of the model.
func (m *Model[ModelType]) String() string {
	return fmt.Sprintf("Model[%s] %+v", reflect.TypeOf(m.M).Name(), m.M)
}

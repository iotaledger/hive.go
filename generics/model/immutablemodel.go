package model

import (
	"context"
	"fmt"
	"reflect"

	"github.com/iotaledger/hive.go/serix"
)

// Model is the base type for all models. It should be embedded in a wrapper type.
// It provides locking and serialization primitives.
type Immutable[ModelType any] struct {
	M ModelType `serix:"0"`
}

// NewImmutable creates a new immutable model instance.
func NewImmutable[ModelType any](model ModelType) (newInstance Model[ModelType]) {
	newInstance = Model[ModelType]{
		M: model,
	}

	return newInstance
}

// FromBytes deserializes a model from a byte slice.
func (m *Immutable[ModelType]) FromBytes(bytes []byte) (err error) {
	_, err = serix.DefaultAPI.Decode(context.Background(), bytes, &m.M, serix.WithValidation())
	return
}

// Bytes serializes a model to a byte slice.
func (m *Immutable[ModelType]) Bytes() (bytes []byte, err error) {
	return serix.DefaultAPI.Encode(context.Background(), m.M, serix.WithValidation())
}

// String returns a string representation of the model.
func (m *Immutable[ModelType]) String() string {
	return fmt.Sprintf("Model[%s] %+v", reflect.TypeOf(m.M).Name(), m.M)
}

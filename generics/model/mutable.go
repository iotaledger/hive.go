package model

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"

	"github.com/iotaledger/hive.go/cerrors"
	"github.com/iotaledger/hive.go/serix"
)

// Mutable is the base type for simple mutable models. It should be embedded in a wrapper type.
// It provides serialization and locking primitives.
type Mutable[OuterModelType any, OuterModelPtrType PtrType[OuterModelType, InnerModelType], InnerModelType any] struct {
	M InnerModelType `serix:"0"`

	sync.RWMutex
}

// NewMutable creates a new mutable model instance.
func NewMutable[OuterModelType any, OuterModelPtrType PtrType[OuterModelType, InnerModelType], InnerModelType any](model *InnerModelType) (newInstance *OuterModelType) {
	newInstance = new(OuterModelType)
	(OuterModelPtrType)(newInstance).New(model)

	return newInstance
}

// New initializes the model with the necessary values when being manually created through a constructor.
func (m *Mutable[OuterModelType, OuterModelPtrType, InnerModelType]) New(innerModel *InnerModelType) {
	m.M = *innerModel
}

// Init initializes the model after it has been restored from it's serialized version.
func (m *Mutable[OuterModelType, OuterModelPtrType, InnerModelType]) Init() {}

// InnerModel returns the inner Model that holds the data.
func (m *Mutable[OuterModelType, OuterModelPtrType, InnerModelType]) InnerModel() *InnerModelType {
	return &m.M
}

// String returns a string representation of the model.
func (m *Mutable[OuterModelType, OuterModelPtrType, InnerModelType]) String() string {
	m.RLock()
	defer m.RUnlock()

	var outerModel OuterModelType

	return fmt.Sprintf(
		"%s {\n\t%s\n}",
		reflect.TypeOf(outerModel).Name(),
		strings.TrimRight(strings.TrimLeft(fmt.Sprintf("%+v", m.M), "{"), "}"),
	)
}

// FromBytes deserializes a model from a byte slice.
func (m *Mutable[OuterModelType, OuterModelPtrType, InnerModelType]) FromBytes(bytes []byte) (err error) {
	m.Init()

	outerInstance := new(OuterModelType)
	consumedBytes, err := serix.DefaultAPI.Decode(context.Background(), bytes, outerInstance, serix.WithValidation())
	if err != nil {
		return errors.Errorf("could not deserialize model: %w", err)
	}
	if len(bytes) != consumedBytes {
		return errors.Errorf("consumed bytes %d not equal total bytes %d: %w", consumedBytes, len(bytes), cerrors.ErrParseBytesFailed)
	}

	m.M = *(OuterModelPtrType)(outerInstance).InnerModel()

	return nil
}

// Bytes serializes a model to a byte slice.
func (m *Mutable[OuterModelType, OuterModelPtrType, InnerModelType]) Bytes() (bytes []byte, err error) {
	m.RLock()
	defer m.RUnlock()

	outerInstance := new(OuterModelType)
	(OuterModelPtrType)(outerInstance).New(&m.M)

	return serix.DefaultAPI.Encode(context.Background(), outerInstance, serix.WithValidation())
}

// Encode serializes the "content of the model" to a byte slice.
func (m *Mutable[OuterModelType, OuterModelPtrType, InnerModelType]) Encode() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()

	return serix.DefaultAPI.Encode(context.Background(), m.M, serix.WithValidation())
}

// Decode deserializes the model from a byte slice.
func (m *Mutable[OuterModelType, OuterModelPtrType, InnerModelType]) Decode(b []byte) (int, error) {
	m.Init()

	return serix.DefaultAPI.Decode(context.Background(), b, &m.M, serix.WithValidation())
}

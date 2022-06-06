package model

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/iotaledger/hive.go/cerrors"
	"github.com/iotaledger/hive.go/serix"
)

// Immutable is the base type for all immutable models. It should be embedded in a wrapper type.
// It provides serialization primitives.
type Immutable[OuterModelType any, OuterModelPtrType PtrType[OuterModelType, InnerModelType], InnerModelType any] struct {
	M InnerModelType
}

// NewImmutable creates a new immutable model instance.
func NewImmutable[OuterModelType any, OuterModelPtrType PtrType[OuterModelType, InnerModelType], InnerModelType any](model *InnerModelType) (newInstance *OuterModelType) {
	newInstance = new(OuterModelType)
	(OuterModelPtrType)(newInstance).New(model)

	return newInstance
}

// New initializes the model with the necessary values when being manually created through a constructor.
func (i *Immutable[OuterModelType, OuterModelPtrType, InnerModelType]) New(innerModel *InnerModelType) {
	i.M = *innerModel
}

// Init initializes the model after it has been restored from it's serialized version.
func (i *Immutable[OuterModelType, OuterModelPtrType, InnerModelType]) Init() {}

// InnerModel returns the inner Model that holds the data.
func (i *Immutable[OuterModelType, OuterModelPtrType, InnerModelType]) InnerModel() *InnerModelType {
	return &i.M
}

// String returns a string representation of the model.
func (i *Immutable[OuterModelType, OuterModelPtrType, InnerModelType]) String() string {
	var outerModel OuterModelType

	return fmt.Sprintf(
		"%s {\n\t%s\n}",
		reflect.TypeOf(outerModel).Name(),
		strings.TrimRight(strings.TrimLeft(fmt.Sprintf("%+v", i.M), "{"), "}"),
	)
}

// FromBytes deserializes a model from a byte slice.
func (i *Immutable[OuterModelType, OuterModelPtrType, InnerModelType]) FromBytes(bytes []byte) (err error) {
	i.Init()

	outerInstance := new(OuterModelType)
	consumedBytes, err := serix.DefaultAPI.Decode(context.Background(), bytes, outerInstance, serix.WithValidation())
	if err != nil {
		return errors.Errorf("could not deserialize model: %w", err)
	}
	if len(bytes) != consumedBytes {
		return errors.Errorf("consumed bytes %d not equal total bytes %d: %w", consumedBytes, len(bytes), cerrors.ErrParseBytesFailed)
	}

	i.M = *(OuterModelPtrType)(outerInstance).InnerModel()

	return nil
}

// Bytes serializes a model to a byte slice.
func (i *Immutable[OuterModelType, OuterModelPtrType, InnerModelType]) Bytes() (bytes []byte, err error) {
	outerInstance := new(OuterModelType)
	(OuterModelPtrType)(outerInstance).New(&i.M)

	return serix.DefaultAPI.Encode(context.Background(), outerInstance, serix.WithValidation())
}

// Encode serializes the "content of the model" to a byte slice.
func (i *Immutable[OuterModelType, OuterModelPtrType, InnerModelType]) Encode() ([]byte, error) {
	return serix.DefaultAPI.Encode(context.Background(), i.M, serix.WithValidation())
}

// Decode deserializes the model from a byte slice.
func (i *Immutable[OuterModelType, OuterModelPtrType, InnerModelType]) Decode(b []byte) (int, error) {
	i.Init()

	return serix.DefaultAPI.Decode(context.Background(), b, &i.M, serix.WithValidation())
}

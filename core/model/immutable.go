package model

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"

	"github.com/izuc/zipp.foundation/serializer/v2/serix"
)

// Immutable is the base type for all immutable models. It should be embedded in a wrapper type.
// It provides serialization primitives.
type Immutable[OuterModelType any, OuterModelPtrType PtrType[OuterModelType, InnerModelType], InnerModelType any] struct {
	M InnerModelType

	bytes      *[]byte // using a pointer here because serix uses reflection which creates a copy of the object
	cacheBytes bool
	bytesMutex *sync.RWMutex
}

// NewImmutable creates a new immutable model instance.
func NewImmutable[OuterModelType any, OuterModelPtrType PtrType[OuterModelType, InnerModelType], InnerModelType any](model *InnerModelType, cacheBytes ...bool) (newInstance *OuterModelType) {
	newInstance = new(OuterModelType)
	(OuterModelPtrType)(newInstance).New(model, cacheBytes...)

	return newInstance
}

// New initializes the model with the necessary values when being manually created through a constructor.
func (i *Immutable[OuterModelType, OuterModelPtrType, InnerModelType]) New(innerModelType *InnerModelType, cacheBytes ...bool) {
	i.Init()

	i.M = *innerModelType

	i.cacheBytes = true
	if len(cacheBytes) > 0 {
		i.cacheBytes = cacheBytes[0]
	}
}

// Init initializes the model after it has been restored from it's serialized version.
func (i *Immutable[OuterModelType, OuterModelPtrType, InnerModelType]) Init() {
	i.cacheBytes = true
	i.bytesMutex = new(sync.RWMutex)
	i.bytes = new([]byte)
}

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
func (i *Immutable[OuterModelType, OuterModelPtrType, InnerModelType]) FromBytes(bytes []byte) (consumedBytes int, err error) {
	i.Init()

	outerInstance := new(OuterModelType)

	if consumedBytes, err = serix.DefaultAPI.Decode(context.Background(), bytes, outerInstance, serix.WithValidation()); err != nil {
		return consumedBytes, errors.Errorf("could not deserialize model: %w", err)
	}
	if len(bytes) != consumedBytes {
		return consumedBytes, errors.Errorf("consumed bytes %d not equal total bytes %d: %w", consumedBytes, len(bytes), ErrParseBytesFailed)
	}

	i.M = *(OuterModelPtrType)(outerInstance).InnerModel()

	if i.cacheBytes {
		i.bytesMutex.Lock()
		defer i.bytesMutex.Unlock()
		// Store the bytes we decoded to avoid any future Encode calls.
		*i.bytes = bytes[:consumedBytes]
	}

	return consumedBytes, nil
}

// Bytes serializes a model to a byte slice.
func (i *Immutable[OuterModelType, OuterModelPtrType, InnerModelType]) Bytes() (bytes []byte, err error) {
	// Return the encoded bytes if we already encoded this object to bytes or decoded it from bytes.
	i.bytesMutex.RLock()
	if i.cacheBytes && i.bytes != nil && len(*i.bytes) > 0 {
		defer i.bytesMutex.RUnlock()

		return *i.bytes, nil
	}
	i.bytesMutex.RUnlock()

	outerInstance := new(OuterModelType)
	(OuterModelPtrType)(outerInstance).New(&i.M)

	encodedBytes, err := serix.DefaultAPI.Encode(context.Background(), outerInstance, serix.WithValidation())
	if err != nil {
		return nil, err
	}

	if i.cacheBytes {
		i.bytesMutex.Lock()
		defer i.bytesMutex.Unlock()

		// Store the encoded bytes to avoid future calls to Encode.
		*i.bytes = encodedBytes
	}

	return encodedBytes, err
}

// Encode serializes the "content of the model" to a byte slice.
func (i Immutable[OuterModelType, OuterModelPtrType, InnerModelType]) Encode() ([]byte, error) {
	return serix.DefaultAPI.Encode(context.Background(), i.M, serix.WithValidation())
}

// Decode deserializes the model from a byte slice.
func (i *Immutable[OuterModelType, OuterModelPtrType, InnerModelType]) Decode(b []byte) (int, error) {
	i.Init()

	return serix.DefaultAPI.Decode(context.Background(), b, &i.M, serix.WithValidation())
}

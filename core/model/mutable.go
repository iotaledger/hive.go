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

// Mutable is the base type for simple mutable models. It should be embedded in a wrapper type.
// It provides serialization and locking primitives.
type Mutable[OuterModelType any, OuterModelPtrType PtrType[OuterModelType, InnerModelType], InnerModelType any] struct {
	M InnerModelType

	bytes      *[]byte // using a pointer here because serix uses reflection which creates a copy of the object
	cacheBytes bool
	bytesMutex *sync.RWMutex

	*sync.RWMutex
}

// NewMutable creates a new mutable model instance.
func NewMutable[OuterModelType any, OuterModelPtrType PtrType[OuterModelType, InnerModelType], InnerModelType any](model *InnerModelType, cacheBytes ...bool) (newInstance *OuterModelType) {
	newInstance = new(OuterModelType)
	(OuterModelPtrType)(newInstance).New(model, cacheBytes...)

	return newInstance
}

// New initializes the model with the necessary values when being manually created through a constructor.
func (m *Mutable[OuterModelType, OuterModelPtrType, InnerModelType]) New(innerModelType *InnerModelType, cacheBytes ...bool) {
	m.Init()

	m.M = *innerModelType

	m.cacheBytes = false
	if len(cacheBytes) > 0 {
		m.cacheBytes = cacheBytes[0]
	}
}

// Init initializes the model after it has been restored from it's serialized version.
func (m *Mutable[OuterModelType, OuterModelPtrType, InnerModelType]) Init() {
	m.RWMutex = new(sync.RWMutex)
	m.bytesMutex = new(sync.RWMutex)
	m.bytes = new([]byte)
}

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

// InvalidateBytesCache invalidates the bytes cache.
func (m *Mutable[OuterModelType, OuterModelPtrType, InnerModelType]) InvalidateBytesCache() {
	m.bytesMutex.Lock()
	defer m.bytesMutex.Unlock()

	m.bytes = new([]byte)
}

// FromBytes deserializes a model from a byte slice.
func (m *Mutable[OuterModelType, OuterModelPtrType, InnerModelType]) FromBytes(bytes []byte) (consumedBytes int, err error) {
	m.Init()

	outerInstance := new(OuterModelType)

	if consumedBytes, err = serix.DefaultAPI.Decode(context.Background(), bytes, outerInstance, serix.WithValidation()); err != nil {
		return consumedBytes, errors.Errorf("could not deserialize model: %w", err)
	}
	if len(bytes) != consumedBytes {
		return consumedBytes, errors.Errorf("consumed bytes %d not equal total bytes %d: %w", consumedBytes, len(bytes), ErrParseBytesFailed)
	}

	m.M = *(OuterModelPtrType)(outerInstance).InnerModel()

	if m.cacheBytes {
		m.bytesMutex.Lock()
		defer m.bytesMutex.Unlock()

		// Store the bytes we decoded to avoid any future Encode calls.
		*m.bytes = bytes[:consumedBytes]
	}

	return consumedBytes, nil
}

// Bytes serializes a model to a byte slice.
func (m *Mutable[OuterModelType, OuterModelPtrType, InnerModelType]) Bytes() (bytes []byte, err error) {
	m.RLock()
	defer m.RUnlock()

	// Return the encoded bytes if we already encoded this object to bytes or decoded it from bytes.
	m.bytesMutex.RLock()
	if m.cacheBytes && m.bytes != nil && len(*m.bytes) > 0 {
		defer m.bytesMutex.RUnlock()

		return *m.bytes, nil
	}
	m.bytesMutex.RUnlock()

	outerInstance := new(OuterModelType)
	(OuterModelPtrType)(outerInstance).New(&m.M)

	encodedBytes, err := serix.DefaultAPI.Encode(context.Background(), outerInstance, serix.WithValidation())
	if err != nil {
		return nil, err
	}

	if m.cacheBytes {
		m.bytesMutex.Lock()
		defer m.bytesMutex.Unlock()

		// Store the encoded bytes to avoid future calls to Encode.
		*m.bytes = encodedBytes
	}

	return encodedBytes, err
}

// Encode serializes the "content of the model" to a byte slice.
func (m Mutable[OuterModelType, OuterModelPtrType, InnerModelType]) Encode() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()

	return serix.DefaultAPI.Encode(context.Background(), m.M, serix.WithValidation())
}

// Decode deserializes the model from a byte slice.
func (m *Mutable[OuterModelType, OuterModelPtrType, InnerModelType]) Decode(b []byte) (int, error) {
	m.Init()

	return serix.DefaultAPI.Decode(context.Background(), b, &m.M, serix.WithValidation())
}

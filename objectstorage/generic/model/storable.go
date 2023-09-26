package model

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"

	"github.com/izuc/zipp.foundation/core/model"
	"github.com/izuc/zipp.foundation/lo"
	"github.com/izuc/zipp.foundation/objectstorage"
	"github.com/izuc/zipp.foundation/runtime/syncutils"
	"github.com/izuc/zipp.foundation/serializer/v2/serix"
)

// Storable is the base type for all storable models. It should be embedded in a wrapper type.
// It provides locking and serialization primitives.
// Types that implement interfaces need to override the serialization logic so that the correct interface can be inferred.
type Storable[IDType, OuterModelType any, OuterModelPtrType model.PtrType[OuterModelType, InnerModelType], InnerModelType any] struct {
	id      *IDType
	idMutex *sync.RWMutex
	M       InnerModelType

	bytes      *[]byte // using a pointer here because serix uses reflection which creates a copy of the object
	cacheBytes bool
	bytesMutex *sync.RWMutex

	*syncutils.RWMutexFake
	*objectstorage.StorableObjectFlags
}

// NewStorable creates a new storable model instance.
func NewStorable[IDType, OuterModelType, InnerModelType any, OuterModelPtrType model.PtrType[OuterModelType, InnerModelType]](model *InnerModelType, cacheBytes ...bool) (newInstance *OuterModelType) {
	newInstance = new(OuterModelType)
	(OuterModelPtrType)(newInstance).New(model, cacheBytes...)

	return newInstance
}

// New initializes the model with the necessary values when being manually created through a constructor.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) New(innerModelType *InnerModelType, cacheBytes ...bool) {
	s.Init()

	s.M = *innerModelType

	s.cacheBytes = true

	if len(cacheBytes) > 0 {
		s.cacheBytes = cacheBytes[0]
	}

	s.SetModified()
}

// Init initializes the model after it has been restored from it's serialized version.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) Init() {
	s.id = new(IDType)

	s.idMutex = new(sync.RWMutex)
	s.bytesMutex = new(sync.RWMutex)
	s.bytes = new([]byte)
	s.RWMutexFake = new(syncutils.RWMutexFake)
	s.StorableObjectFlags = new(objectstorage.StorableObjectFlags)

	s.Persist()
}

// InnerModel returns the inner Model that holds the data.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) InnerModel() *InnerModelType {
	return &s.M
}

// ID returns the ID of the model.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) ID() (id IDType) {
	s.idMutex.RLock()
	defer s.idMutex.RUnlock()

	if s.id == nil {
		panic(fmt.Sprintf("ID is not set for %v", s))
	}

	return *s.id
}

// SetID sets the ID of the model.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) SetID(id IDType) {
	s.idMutex.Lock()
	defer s.idMutex.Unlock()

	*s.id = id
}

// String returns a string representation of the model.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) String() string {
	s.RLock()
	defer s.RUnlock()

	var outerModel OuterModelType

	return fmt.Sprintf("%s {\n\tID: %+v\n\t%s\n}", reflect.TypeOf(outerModel).Name(), s.id, strings.TrimRight(strings.TrimLeft(fmt.Sprintf("%+v", s.M), "{"), "}"))
}

// IDFromBytes deserializes an ID from a byte slice.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) IDFromBytes(bytes []byte) (err error) {
	if _, err = serix.DefaultAPI.Decode(context.Background(), bytes, s.id); err != nil {
		return errors.Errorf("failed to read IF from bytes: %w", err)
	}

	return
}

// FromBytes deserializes a model from a byte slice.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) FromBytes(bytes []byte) (consumedBytes int, err error) {
	outerInstance := new(OuterModelType)
	if consumedBytes, err = serix.DefaultAPI.Decode(context.Background(), bytes, outerInstance, serix.WithValidation()); err != nil {
		return consumedBytes, errors.Errorf("could not deserialize model: %w", err)
	}
	if len(bytes) != consumedBytes {
		return consumedBytes, errors.Errorf("consumed bytes %d not equal total bytes %d: %w", consumedBytes, len(bytes), ErrParseBytesFailed)
	}

	s.Init()
	s.M = *(OuterModelPtrType)(outerInstance).InnerModel()

	if s.cacheBytes {
		s.bytesMutex.Lock()
		defer s.bytesMutex.Unlock()
		// Store the bytes we decoded to avoid any future Encode calls.
		*s.bytes = bytes[:consumedBytes]
	}

	return consumedBytes, nil
}

// Bytes serializes a model to a byte slice.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) Bytes() (bytes []byte, err error) {
	s.RLock()
	defer s.RUnlock()

	// Return the encoded bytes if we already encoded this object to bytes or decoded it from bytes.
	s.bytesMutex.RLock()
	if s.cacheBytes && s.bytes != nil && len(*s.bytes) > 0 {
		defer s.bytesMutex.RUnlock()

		return *s.bytes, nil
	}
	s.bytesMutex.RUnlock()

	outerInstance := new(OuterModelType)
	(OuterModelPtrType)(outerInstance).New(&s.M)

	encodedBytes, err := serix.DefaultAPI.Encode(context.Background(), outerInstance, serix.WithValidation())
	if err != nil {
		return nil, err
	}

	if s.cacheBytes {
		s.bytesMutex.Lock()
		defer s.bytesMutex.Unlock()

		// Store the encoded bytes to avoid future calls to Encode.
		*s.bytes = encodedBytes
	}

	return encodedBytes, err
}

// InvalidateBytesCache invalidates the bytes cache.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) InvalidateBytesCache() {
	s.bytesMutex.Lock()
	defer s.bytesMutex.Unlock()

	s.bytes = new([]byte)
}

// FromObjectStorage deserializes a model from the object storage.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) FromObjectStorage(key, data []byte) (err error) {
	if _, err = s.FromBytes(data); err != nil {
		return errors.Errorf("failed to decode Model: %w", err)
	}

	if err = s.IDFromBytes(key); err != nil {
		return errors.Errorf("failed to decode ID: %w", err)
	}

	return nil
}

// ImportModel imports the model of the other storable.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) ImportStorable(other *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) {
	s.Lock()
	defer s.Unlock()

	other.RLock()
	defer other.RUnlock()

	s.bytesMutex.Lock()
	defer s.bytesMutex.Unlock()

	other.bytesMutex.Lock()
	defer other.bytesMutex.Unlock()

	s.M = other.M
	s.bytes = other.bytes
}

// ObjectStorageKey returns the bytes, that are used as a key to store the object in the k/v store.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) ObjectStorageKey() (key []byte) {
	key, err := serix.DefaultAPI.Encode(context.Background(), s.ID(), serix.WithValidation())
	if err != nil {
		panic(err)
	}

	return key
}

// ObjectStorageValue returns the bytes, that are stored in the value part of the k/v store.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) ObjectStorageValue() (value []byte) {
	return lo.PanicOnErr(s.Bytes())
}

// Encode serializes the "content of the model" to a byte slice.
func (s Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) Encode() ([]byte, error) {
	return serix.DefaultAPI.Encode(context.Background(), s.M, serix.WithValidation())
}

// Decode deserializes the model from a byte slice.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) Decode(b []byte) (int, error) {
	s.Init()

	return serix.DefaultAPI.Decode(context.Background(), b, &s.M, serix.WithValidation())
}

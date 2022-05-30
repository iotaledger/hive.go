package model

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/cockroachdb/errors"

	"github.com/iotaledger/hive.go/generics/lo"
	"github.com/iotaledger/hive.go/generics/objectstorage"
	"github.com/iotaledger/hive.go/serix"
)

// Storable is the base type for all storable models. It should be embedded in a wrapper type.
// It provides locking and serialization primitives.
// Types that implement interfaces need to override the serialization logic so that the correct interface can be inferred.
type Storable[IDType, ModelType any] struct {
	id      *IDType
	idFunc  func(model *ModelType) IDType
	idMutex sync.RWMutex
	M       ModelType `serix:"0"`

	sync.RWMutex
	objectstorage.StorableObjectFlags
}

// NewStorable creates a new storable model instance.
func NewStorable[IDType, ModelType any](model ModelType, optionalIDFunc ...func(model *ModelType) IDType) (new Storable[IDType, ModelType]) {
	if len(optionalIDFunc) == 0 {
		optionalIDFunc = append(optionalIDFunc, func(model *ModelType) (key IDType) {
			return key
		})
	}

	new = Storable[IDType, ModelType]{
		M:      model,
		idFunc: optionalIDFunc[0],
	}
	new.SetModified()
	new.Persist()

	return new
}

// ID returns the ID of the model.
func (s *Storable[IDType, ModelType]) ID() (id IDType) {
	s.idMutex.RLock()
	if s.id != nil {
		defer s.idMutex.RUnlock()
		return *s.id
	}
	s.idMutex.RUnlock()

	s.idMutex.Lock()
	defer s.idMutex.Unlock()
	if s.id != nil {
		return *s.id
	}

	id = s.idFunc(&s.M)
	s.id = &id

	return id
}

// SetID sets the ID of the model.
func (s *Storable[IDType, ModelType]) SetID(id IDType) {
	s.idMutex.Lock()
	defer s.idMutex.Unlock()

	s.id = &id
}

// IDFromBytes deserializes an ID from a byte slice.
func (s *Storable[IDType, ModelType]) IDFromBytes(bytes []byte) (err error) {
	s.idMutex.Lock()
	defer s.idMutex.Unlock()

	_, err = serix.DefaultAPI.Decode(context.Background(), bytes, &s.id)
	return
}

// FromBytes deserializes a model from a byte slice.
func (s *Storable[IDType, ModelType]) FromBytes(bytes []byte) (err error) {
	s.Lock()
	defer s.Unlock()

	_, err = serix.DefaultAPI.Decode(context.Background(), bytes, &s.M, serix.WithValidation())
	return
}

// FromObjectStorage deserializes a model stored in the object storage.
func (s *Storable[IDType, ModelType]) FromObjectStorage(key, data []byte) (err error) {
	if err = s.IDFromBytes(key); err != nil {
		return errors.Errorf("failed to decode ID: %w", err)
	}

	if err = s.FromBytes(data); err != nil {
		return errors.Errorf("failed to decode Model: %w", err)
	}

	return nil
}

// ObjectStorageKey returns the bytes, that are used as a key to store the object in the k/v store.
func (s *Storable[IDType, ModelType]) ObjectStorageKey() (key []byte) {
	key, err := serix.DefaultAPI.Encode(context.Background(), s.ID(), serix.WithValidation())
	if err != nil {
		panic(err)
	}

	return key
}

// ObjectStorageValue returns the bytes, that are stored in the value part of the k/v store.
func (s *Storable[IDType, ModelType]) ObjectStorageValue() (value []byte) {
	return lo.PanicOnErr(s.Bytes())
}

// Bytes serializes a model to a byte slice.
func (s *Storable[IDType, ModelType]) Bytes() (bytes []byte, err error) {
	s.RLock()
	defer s.RUnlock()

	return serix.DefaultAPI.Encode(context.Background(), s.M, serix.WithValidation())
}

// String returns a string representation of the model.
func (s *Storable[IDType, ModelType]) String() string {
	return fmt.Sprintf("Storable[%s] {\n\tID: %+v\n\tModel: %+v\n}", reflect.TypeOf(s.M).Name(), s.id, s.M)
}

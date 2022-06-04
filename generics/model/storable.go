package model

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"

	"github.com/iotaledger/hive.go/cerrors"
	"github.com/iotaledger/hive.go/generics/lo"
	"github.com/iotaledger/hive.go/generics/objectstorage"
	"github.com/iotaledger/hive.go/serix"
)

// Storable is the base type for all storable models. It should be embedded in a wrapper type.
// It provides locking and serialization primitives.
// Types that implement interfaces need to override the serialization logic so that the correct interface can be inferred.
type Storable[IDType, OuterModelType any, OuterModelPtrType outerStorableModelPtr[OuterModelType, InnerModelType], InnerModelType any] struct {
	id      *IDType
	idMutex sync.RWMutex

	M InnerModelType `serix:"0"`

	sync.RWMutex
	objectstorage.StorableObjectFlags
}

// NewStorable creates a new storable model instance.
func NewStorable[IDType, OuterModelType, InnerModelType any, OuterModelPtrType outerStorableModelPtr[OuterModelType, InnerModelType]](model *InnerModelType) (newInstance *OuterModelType) {
	newInstance = new(OuterModelType)
	(OuterModelPtrType)(newInstance).Init(model)

	return newInstance
}

func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) Init(innerModel *InnerModelType) {
	s.id = new(IDType)
	s.M = *innerModel

	s.SetModified()
	s.Persist()
}

// IDFromBytes deserializes an ID from a byte slice.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) IDFromBytes(bytes []byte) (err error) {
	s.idMutex.Lock()
	defer s.idMutex.Unlock()

	if s.id == nil {
		s.id = new(IDType)
	}

	if _, err = serix.DefaultAPI.Decode(context.Background(), bytes, s.id); err != nil {
		return errors.Errorf("failed to read IF from bytes: %w", err)
	}

	return
}

// FromBytes deserializes a model from a byte slice.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) FromBytes(bytes []byte) (err error) {
	s.Lock()
	defer s.Unlock()

	outerInstance := new(OuterModelType)

	consumedBytes, err := serix.DefaultAPI.Decode(context.Background(), bytes, outerInstance, serix.WithValidation())
	if err != nil {
		return errors.Errorf("could not deserialize model: %w", err)
	}

	if len(bytes) != consumedBytes {
		return errors.Errorf("consumed bytes %d not equal total bytes %d: %w", consumedBytes, len(bytes), cerrors.ErrParseBytesFailed)
	}

	s.M = *(OuterModelPtrType)(outerInstance).InnerModel()

	return nil
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

// InnerModel returns the inner Model that holds the data.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) InnerModel() *InnerModelType {
	return &s.M
}

// Bytes serializes a model to a byte slice.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) Bytes() (bytes []byte, err error) {
	s.RLock()
	defer s.RUnlock()

	outerInstance := new(OuterModelType)
	(OuterModelPtrType)(outerInstance).Init(&s.M)

	return serix.DefaultAPI.Encode(context.Background(), outerInstance, serix.WithValidation())
}

// String returns a string representation of the model.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) String() string {
	var outerModel OuterModelType

	return fmt.Sprintf(
		"%s {\n\tID: %+v\n\t%s\n}",
		reflect.TypeOf(outerModel).Name(),
		s.id,
		strings.TrimRight(strings.TrimLeft(fmt.Sprintf("%+v", s.M), "{"), "}"),
	)
}

// region object storage interface /////////////////////////////////////////////////////////////////////////////////////

// FromObjectStorage deserializes a model stored in the object storage.
func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) FromObjectStorage(key, data []byte) (err error) {
	if err = s.IDFromBytes(key); err != nil {
		return errors.Errorf("failed to decode ID: %w", err)
	}

	if err = s.FromBytes(data); err != nil {
		return errors.Errorf("failed to decode Model: %w", err)
	}

	return nil
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

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region serix interface //////////////////////////////////////////////////////////////////////////////////////////////

func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) Encode() ([]byte, error) {
	s.RLock()
	defer s.RUnlock()

	return serix.DefaultAPI.Encode(context.Background(), s.M, serix.WithValidation())
}

func (s *Storable[IDType, OuterModelType, OuterModelPtrType, InnerModelType]) Decode(b []byte) (int, error) {
	s.Lock()
	defer s.Unlock()

	return serix.DefaultAPI.Decode(context.Background(), b, &s.M, serix.WithValidation())
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

package model

import (
	"context"
	"fmt"
	"reflect"

	"github.com/cockroachdb/errors"

	"github.com/izuc/zipp.foundation/core/model"
	"github.com/izuc/zipp.foundation/lo"
	"github.com/izuc/zipp.foundation/objectstorage"
	"github.com/izuc/zipp.foundation/serializer/v2/byteutils"
	"github.com/izuc/zipp.foundation/serializer/v2/serix"
)

// StorableReference is the base type for all storable reference models. It should be embedded in a wrapper type.
// It provides locking and serialization primitives.
type StorableReference[OuterModelType any, OuterModelPtrType model.ReferencePtrType[OuterModelType, SourceIDType, TargetIDType], SourceIDType, TargetIDType any] struct {
	sourceID SourceIDType
	targetID TargetIDType

	*objectstorage.StorableObjectFlags
}

// NewStorableReference creates a new storable reference model instance.
func NewStorableReference[OuterModelType any, OuterModelPtrType model.ReferencePtrType[OuterModelType, SourceIDType, TargetIDType], SourceIDType, TargetIDType any](sourceID SourceIDType, targetID TargetIDType) (newInstance *OuterModelType) {
	newInstance = new(OuterModelType)
	(OuterModelPtrType)(newInstance).New(sourceID, targetID)

	return newInstance
}

// New initializes the storable reference model with the necessary values when being manually created through a constructor.
func (s *StorableReference[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType]) New(sourceID SourceIDType, targetID TargetIDType) {
	s.Init()

	s.sourceID = sourceID
	s.targetID = targetID

	s.SetModified()
}

// Init initializes the storable reference model after it has been restored from it's serialized version.
func (s *StorableReference[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType]) Init() {
	s.StorableObjectFlags = new(objectstorage.StorableObjectFlags)

	s.Persist()
}

// SourceID returns the SourceID of the storable reference model.
func (s *StorableReference[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType]) SourceID() SourceIDType {
	return s.sourceID
}

// TargetID returns the TargetID of the storable reference model.
func (s *StorableReference[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType]) TargetID() TargetIDType {
	return s.targetID
}

// String returns a string representation of the model.
func (s *StorableReference[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType]) String() string {
	var outerModel OuterModelType

	return fmt.Sprintf("%s {\n\tSourceID: %+v\n\tTargetID: %+v\n}", reflect.TypeOf(outerModel).Name(), s.sourceID, s.targetID)
}

// FromBytes deserializes a storable reference model from a byte slice.
func (s *StorableReference[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType]) FromBytes(bytes []byte) (consumedBytes int, err error) {
	outerInstance := new(OuterModelType)

	if consumedBytes, err = serix.DefaultAPI.Decode(context.Background(), bytes, outerInstance, serix.WithValidation()); err != nil {
		return consumedBytes, errors.Errorf("could not deserialize reference: %w", err)
	}
	if len(bytes) != consumedBytes {
		return consumedBytes, errors.Errorf("consumed bytes %d not equal total bytes %d: %w", consumedBytes, len(bytes), ErrParseBytesFailed)
	}

	s.Init()

	s.sourceID = (OuterModelPtrType)(outerInstance).SourceID()
	s.targetID = (OuterModelPtrType)(outerInstance).TargetID()

	return consumedBytes, nil
}

// Bytes serializes a storable reference model to a byte slice.
func (s *StorableReference[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType]) Bytes() (bytes []byte, err error) {
	outerInstance := new(OuterModelType)
	(OuterModelPtrType)(outerInstance).New(s.sourceID, s.targetID)

	return serix.DefaultAPI.Encode(context.Background(), outerInstance, serix.WithValidation())
}

// FromObjectStorage deserializes a model from the object storage.
func (s *StorableReference[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType]) FromObjectStorage(key, _ []byte) (err error) {
	if _, err = s.Decode(key); err != nil {
		return errors.Errorf("failed to decode storable reference: %w", err)
	}

	return nil
}

// ObjectStorageKey returns the bytes, that are used as a key to store the object in the k/v store.
func (s *StorableReference[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType]) ObjectStorageKey() (key []byte) {
	return lo.PanicOnErr(s.Encode())
}

// ObjectStorageValue returns the bytes, that are stored in the value part of the k/v store.
func (s *StorableReference[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType]) ObjectStorageValue() (value []byte) {
	return nil
}

// KeyPartitions returns a slice of the key partitions that are used to store the object in the k/v store.
func (s *StorableReference[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType]) KeyPartitions() []int {
	var sourceID SourceIDType
	var targetID TargetIDType

	return []int{len(lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), sourceID))), len(lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), targetID)))}
}

// Encode serializes the "content of the model" to a byte slice.
func (s StorableReference[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType]) Encode() ([]byte, error) {
	sourceIDBytes, err := serix.DefaultAPI.Encode(context.Background(), s.sourceID, serix.WithValidation())
	if err != nil {
		return nil, errors.Errorf("could not encode source id: %w", err)
	}

	targetIDBytes, err := serix.DefaultAPI.Encode(context.Background(), s.targetID, serix.WithValidation())
	if err != nil {
		return nil, errors.Errorf("could not encode target id: %w", err)
	}

	return byteutils.ConcatBytes(sourceIDBytes, targetIDBytes), nil
}

// Decode deserializes the model from a byte slice.
func (s *StorableReference[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType]) Decode(b []byte) (int, error) {
	s.Init()

	consumedSourceIDBytes, err := serix.DefaultAPI.Decode(context.Background(), b, &s.sourceID, serix.WithValidation())
	if err != nil {
		return 0, errors.Errorf("could not decode source id: %w", err)
	}

	consumedTargetIDBytes, err := serix.DefaultAPI.Decode(context.Background(), b[consumedSourceIDBytes:], &s.targetID, serix.WithValidation())
	if err != nil {
		return 0, errors.Errorf("could not decode target id: %w", err)
	}

	return consumedSourceIDBytes + consumedTargetIDBytes, nil
}

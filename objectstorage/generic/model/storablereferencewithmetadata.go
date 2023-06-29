package model

import (
	"context"
	"fmt"
	"reflect"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/objectstorage"
	"github.com/iotaledger/hive.go/runtime/syncutils"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	"github.com/iotaledger/hive.go/serializer/v2/serix/model"
)

// StorableReferenceWithMetadata is the base type for all storable reference models. It should be embedded in a wrapper type.
// It provides locking and serialization primitives.
type StorableReferenceWithMetadata[OuterModelType any, OuterModelPtrType model.ReferenceWithMetadataPtrType[OuterModelType, SourceIDType, TargetIDType, InnerModelType], SourceIDType, TargetIDType, InnerModelType any] struct {
	sourceID SourceIDType
	targetID TargetIDType
	M        InnerModelType

	*syncutils.RWMutexFake
	*objectstorage.StorableObjectFlags
}

// NewStorableReferenceWithMetadata creates a new storable reference model instance.
func NewStorableReferenceWithMetadata[OuterModelType any, OuterModelPtrType model.ReferenceWithMetadataPtrType[OuterModelType, SourceIDType, TargetIDType, InnerModelType], SourceIDType, TargetIDType, InnerModelType any](sourceID SourceIDType, targetID TargetIDType, model *InnerModelType) (newInstance *OuterModelType) {
	newInstance = new(OuterModelType)
	(OuterModelPtrType)(newInstance).New(sourceID, targetID, model)

	return newInstance
}

// New initializes the storable reference model with the necessary values when being manually created through a constructor.
func (s *StorableReferenceWithMetadata[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType, InnerModelType]) New(sourceID SourceIDType, targetID TargetIDType, model *InnerModelType) {
	s.Init()

	s.sourceID = sourceID
	s.targetID = targetID
	s.M = *model

	s.SetModified()
}

// Init initializes the storable reference model after it has been restored from it's serialized version.
func (s *StorableReferenceWithMetadata[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType, InnerModelType]) Init() {
	s.StorableObjectFlags = new(objectstorage.StorableObjectFlags)
	s.RWMutexFake = new(syncutils.RWMutexFake)

	s.Persist()
}

// SourceID returns the SourceID of the storable reference model.
func (s *StorableReferenceWithMetadata[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType, InnerModelType]) SourceID() SourceIDType {
	return s.sourceID
}

// TargetID returns the TargetID of the storable reference model.
func (s *StorableReferenceWithMetadata[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType, InnerModelType]) TargetID() TargetIDType {
	return s.targetID
}

// InnerModel returns the inner Model that holds the data.
func (s *StorableReferenceWithMetadata[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType, InnerModelType]) InnerModel() *InnerModelType {
	return &s.M
}

// String returns a string representation of the model.
func (s *StorableReferenceWithMetadata[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType, InnerModelType]) String() string {
	s.RLock()
	defer s.RUnlock()

	var outerModel OuterModelType

	return fmt.Sprintf("%s {\n\tSourceID: %+v\n\tTargetID: %+v\n\tModel: %+v\n}", reflect.TypeOf(outerModel).Name(), s.sourceID, s.targetID, s.M)
}

// FromBytes deserializes a storable reference model from a byte slice.
func (s *StorableReferenceWithMetadata[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType, InnerModelType]) FromBytes(bytes []byte) (consumedBytes int, err error) {
	outerInstance := new(OuterModelType)

	if consumedBytes, err = serix.DefaultAPI.Decode(context.Background(), bytes, outerInstance, serix.WithValidation()); err != nil {
		return consumedBytes, ierrors.Wrap(err, "could not deserialize reference")
	}
	if len(bytes) != consumedBytes {
		return consumedBytes, ierrors.Wrapf(ErrParseBytesFailed, "consumed bytes %d not equal total bytes %d", consumedBytes, len(bytes))
	}

	s.Init()
	s.sourceID = (OuterModelPtrType)(outerInstance).SourceID()
	s.targetID = (OuterModelPtrType)(outerInstance).TargetID()
	s.M = *(OuterModelPtrType)(outerInstance).InnerModel()

	return consumedBytes, nil
}

// Bytes serializes a storable reference model to a byte slice.
func (s *StorableReferenceWithMetadata[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType, InnerModelType]) Bytes() (bytes []byte, err error) {
	s.RLock()
	defer s.RUnlock()

	outerInstance := new(OuterModelType)
	(OuterModelPtrType)(outerInstance).New(s.sourceID, s.targetID, &s.M)

	return serix.DefaultAPI.Encode(context.Background(), outerInstance, serix.WithValidation())
}

// FromObjectStorage deserializes a model from the object storage.
func (s *StorableReferenceWithMetadata[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType, InnerModelType]) FromObjectStorage(key, value []byte) (err error) {
	if _, err = s.Decode(byteutils.ConcatBytes(key, value)); err != nil {
		return ierrors.Wrap(err, "failed to decode storable reference")
	}

	return nil
}

// ObjectStorageKey returns the bytes, that are used as a key to store the object in the k/v store.
func (s *StorableReferenceWithMetadata[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType, InnerModelType]) ObjectStorageKey() (key []byte) {
	sourceIDBytes, err := serix.DefaultAPI.Encode(context.Background(), s.SourceID, serix.WithValidation())
	if err != nil {
		panic(ierrors.Wrap(err, "could not encode source id"))
	}

	targetIDBytes, err := serix.DefaultAPI.Encode(context.Background(), s.TargetID, serix.WithValidation())
	if err != nil {
		panic(ierrors.Wrap(err, "could not encode target id"))
	}

	return byteutils.ConcatBytes(sourceIDBytes, targetIDBytes)
}

// ObjectStorageValue returns the bytes, that are stored in the value part of the k/v store.
func (s *StorableReferenceWithMetadata[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType, InnerModelType]) ObjectStorageValue() (value []byte) {
	s.RLock()
	defer s.RUnlock()

	modelBytes, err := serix.DefaultAPI.Encode(context.Background(), s.M, serix.WithValidation())
	if err != nil {
		panic(ierrors.Wrap(err, "could not encode model"))
	}

	return modelBytes
}

// KeyPartitions returns a slice of the key partitions that are used to store the object in the k/v store.
func (s *StorableReferenceWithMetadata[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType, InnerModelType]) KeyPartitions() []int {
	var sourceID SourceIDType
	var targetID TargetIDType

	return []int{len(lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), sourceID))), len(lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), targetID)))}
}

// Encode serializes the "content of the model" to a byte slice.
func (s StorableReferenceWithMetadata[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType, InnerModelType]) Encode() ([]byte, error) {
	sourceIDBytes, err := serix.DefaultAPI.Encode(context.Background(), s.SourceID, serix.WithValidation())
	if err != nil {
		return nil, ierrors.Wrap(err, "could not encode source id")
	}

	targetIDBytes, err := serix.DefaultAPI.Encode(context.Background(), s.TargetID, serix.WithValidation())
	if err != nil {
		return nil, ierrors.Wrap(err, "could not encode target id")
	}

	s.RLock()
	defer s.RUnlock()

	modelBytes, err := serix.DefaultAPI.Encode(context.Background(), s.M, serix.WithValidation())
	if err != nil {
		return nil, ierrors.Wrap(err, "could not encode model")
	}

	return byteutils.ConcatBytes(sourceIDBytes, targetIDBytes, modelBytes), nil
}

// Decode deserializes the model from a byte slice.
func (s *StorableReferenceWithMetadata[OuterModelType, OuterModelPtrType, SourceIDType, TargetIDType, InnerModelType]) Decode(b []byte) (int, error) {
	s.Init()

	consumedSourceIDBytes, err := serix.DefaultAPI.Decode(context.Background(), b, &s.sourceID, serix.WithValidation())
	if err != nil {
		return 0, ierrors.Wrap(err, "could not decode source id")
	}

	consumedTargetIDBytes, err := serix.DefaultAPI.Decode(context.Background(), b[consumedSourceIDBytes:], &s.targetID, serix.WithValidation())
	if err != nil {
		return 0, ierrors.Wrap(err, "could not decode target id")
	}

	consumedModelIDBytes, err := serix.DefaultAPI.Decode(context.Background(), b[consumedSourceIDBytes+consumedTargetIDBytes:], &s.M, serix.WithValidation())
	if err != nil {
		return 0, ierrors.Wrap(err, "could not decode model")
	}

	return consumedSourceIDBytes + consumedTargetIDBytes + consumedModelIDBytes, nil
}

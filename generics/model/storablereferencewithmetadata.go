package model

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/cockroachdb/errors"

	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/generics/lo"
	"github.com/iotaledger/hive.go/objectstorage"
	"github.com/iotaledger/hive.go/serix"
)

// StorableReferenceWithMetadata is the base type for all storable reference models with metadata. It should be embedded
// in a wrapper type. It provides locking and serialization primitives.
type StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType any] struct {
	SourceID SourceIDType
	TargetID TargetIDType
	M        ModelType `serix:"0"`

	sync.RWMutex
	objectstorage.StorableObjectFlags
}

// NewStorableReferenceWithMetadata creates a new storable reference with metadata model instance.
func NewStorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType any](s SourceIDType, t TargetIDType, m ModelType) (new StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType]) {
	new = StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType]{
		SourceID: s,
		TargetID: t,
		M:        m,
	}
	new.SetModified()
	new.Persist()

	return new
}

// FromBytes deserializes a model from a byte slice.
func (s *StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType]) FromBytes(bytes []byte) (err error) {
	s.Lock()
	defer s.Unlock()

	consumedSourceIDBytes, err := serix.DefaultAPI.Decode(context.Background(), bytes, &s.SourceID, serix.WithValidation())
	if err != nil {
		return errors.Errorf("failed to decode source ID: %w", err)
	}
	consumedTargetIDBytes, err := serix.DefaultAPI.Decode(context.Background(), bytes[consumedSourceIDBytes:], &s.TargetID, serix.WithValidation())
	if err != nil {
		return errors.Errorf("failed to decode target ID: %w", err)
	}
	if _, err = serix.DefaultAPI.Decode(context.Background(), bytes[consumedSourceIDBytes+consumedTargetIDBytes:], &s.M, serix.WithValidation()); err != nil {
		return errors.Errorf("failed to decode model: %w", err)
	}

	return nil
}

// FromObjectStorage deserializes a model stored in the object storage.
func (s *StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType]) FromObjectStorage(key, data []byte) (err error) {
	if err = s.FromBytes(byteutils.ConcatBytes(key, data)); err != nil {
		err = errors.Errorf("failed to load object from object storage: %w", err)
	}

	return
}

// ObjectStorageKey returns the bytes, that are used as a key to store the object in the k/v store.
func (s *StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType]) ObjectStorageKey() (key []byte) {
	s.RLock()
	defer s.RUnlock()

	return byteutils.ConcatBytes(lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), s.SourceID)), lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), s.TargetID)))
}

// ObjectStorageValue returns the bytes, that are stored in the value part of the k/v store.
func (s *StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType]) ObjectStorageValue() (value []byte) {
	s.RLock()
	defer s.RUnlock()

	return lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), &s.M))
}

// KeyPartitions returns a slice of the key partitions that are used to store the object in the k/v store.
func (s StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType]) KeyPartitions() []int {
	var sourceID SourceIDType
	var targetID TargetIDType

	return []int{len(lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), sourceID))), len(lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), targetID)))}
}

// Bytes serializes a model to a byte slice.
func (s *StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType]) Bytes() (bytes []byte, err error) {
	s.RLock()
	defer s.RUnlock()

	sourceIDBytes, err := serix.DefaultAPI.Encode(context.Background(), s.SourceID)
	if err != nil {
		return nil, errors.Errorf("failed to serialize source ID: %w", err)
	}
	targetIDBytes, err := serix.DefaultAPI.Encode(context.Background(), s.TargetID)
	if err != nil {
		return nil, errors.Errorf("failed to serialize target ID: %w", err)
	}
	modelBytes, err := serix.DefaultAPI.Encode(context.Background(), s.M)
	if err != nil {
		return nil, errors.Errorf("failed to serialize model: %w", err)
	}

	return byteutils.ConcatBytes(sourceIDBytes, targetIDBytes, modelBytes), nil
}

// String returns a string representation of the model.
func (s *StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType]) String() string {
	s.RLock()
	defer s.RUnlock()

	return fmt.Sprintf("StorableReferenceWithMetadata[%s, %s, %s] {\n\tSourceID: %+v\n\tTargetID: %+v\n\tModel: %+v\n}",
		reflect.TypeOf(s.SourceID).Name(), reflect.TypeOf(s.TargetID).Name(), reflect.TypeOf(s.M).Name(), s.SourceID, s.TargetID, s.M)
}

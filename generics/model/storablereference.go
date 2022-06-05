package model

import (
	"context"
	"fmt"
	"reflect"

	"github.com/cockroachdb/errors"

	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/generics/lo"
	"github.com/iotaledger/hive.go/generics/objectstorage"
	"github.com/iotaledger/hive.go/serix"
)

// StorableReference is the base type for all storable reference models. It should be embedded in a wrapper type.
// It provides locking and serialization primitives.
type StorableReference[SourceIDType, TargetIDType any] struct {
	SourceID SourceIDType
	TargetID TargetIDType

	objectstorage.StorableObjectFlags
}

// NewStorableReference creates a new storable reference model instance.
func NewStorableReference[SourceIDType, TargetIDType any](s SourceIDType, t TargetIDType) (new StorableReference[SourceIDType, TargetIDType]) {
	new = StorableReference[SourceIDType, TargetIDType]{
		SourceID: s,
		TargetID: t,
	}
	new.SetModified()
	new.Persist()

	return new
}

// FromBytes deserializes a model from a byte slice.
func (s *StorableReference[SourceIDType, TargetIDType]) FromBytes(bytes []byte) (err error) {
	consumedBytesSource, err := serix.DefaultAPI.Decode(context.Background(), bytes, &s.SourceID, serix.WithValidation())
	if err != nil {
		return errors.Errorf("failed to decode SourceID: %w", err)
	}

	_, err = serix.DefaultAPI.Decode(context.Background(), bytes[consumedBytesSource:], &s.TargetID, serix.WithValidation())
	if err != nil {
		return errors.Errorf("failed to decode TargetID: %w", err)
	}

	return err
}

// FromObjectStorage deserializes a model stored in the object storage.
func (s *StorableReference[SourceIDType, TargetIDType]) FromObjectStorage(key, _ []byte) (err error) {
	if err = s.FromBytes(key); err != nil {
		err = errors.Errorf("failed to load object from object storage: %w", err)
	}

	return
}

// ObjectStorageKey returns the bytes, that are used as a key to store the object in the k/v store.
func (s *StorableReference[SourceIDType, TargetIDType]) ObjectStorageKey() (key []byte) {
	return lo.PanicOnErr(s.Bytes())
}

// ObjectStorageValue returns the bytes, that are stored in the value part of the k/v store. For a storable reference
// the value is nil.
func (s *StorableReference[SourceIDType, TargetIDType]) ObjectStorageValue() (value []byte) {
	return nil
}

// KeyPartitions returns a slice of the key partitions that are used to store the object in the k/v store.
func (s StorableReference[SourceIDType, TargetIDType]) KeyPartitions() []int {
	var sourceID SourceIDType
	var targetID TargetIDType

	return []int{len(lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), sourceID))), len(lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), targetID)))}
}

// Bytes serializes a model to a byte slice.
func (s *StorableReference[SourceIDType, TargetIDType]) Bytes() (bytes []byte, err error) {
	sourceBytes, err := serix.DefaultAPI.Encode(context.Background(), s.SourceID, serix.WithValidation())
	if err != nil {
		return nil, errors.Errorf("failed to encode source ID: %w", err)
	}
	targetBytes, err := serix.DefaultAPI.Encode(context.Background(), s.TargetID, serix.WithValidation())
	if err != nil {
		return nil, errors.Errorf("failed to encode target ID: %w", err)
	}

	return byteutils.ConcatBytes(sourceBytes, targetBytes), nil
}

// String returns a string representation of the model.
func (s *StorableReference[SourceIDType, TargetIDType]) String() string {
	return fmt.Sprintf("StorableReference[%s, %s] {\n\tSourceID: %+v\n\tTargetID: %+v\n}",
		reflect.TypeOf(s.SourceID).Name(), reflect.TypeOf(s.TargetID).Name(), s.SourceID, s.TargetID)
}

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

type StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType any] struct {
	SourceID SourceIDType
	TargetID TargetIDType
	M        ModelType `serix:"0"`

	sync.RWMutex
	objectstorage.StorableObjectFlags
}

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

func (s *StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType]) FromObjectStorage(key, data []byte) (err error) {
	if err = s.FromBytes(byteutils.ConcatBytes(key, data)); err != nil {
		err = errors.Errorf("failed to load object from object storage: %w", err)
	}

	return
}

func (s *StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType]) ObjectStorageKey() (key []byte) {
	s.RLock()
	defer s.RUnlock()

	return byteutils.ConcatBytes(lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), s.SourceID)), lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), s.TargetID)))
}

func (s *StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType]) ObjectStorageValue() (value []byte) {
	s.RLock()
	defer s.RUnlock()

	return lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), &s.M))
}

func (s StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType]) KeyPartitions() []int {
	var sourceID SourceIDType
	var targetID TargetIDType

	return []int{len(lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), sourceID))), len(lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), targetID)))}
}

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

func (s *StorableReferenceWithMetadata[SourceIDType, TargetIDType, ModelType]) String() string {
	s.RLock()
	defer s.RUnlock()

	return fmt.Sprintf("StorableReferenceWithMetadata[%s, %s, %s] {\n\tSourceID: %+v\n\tTargetID: %+v\n\tModel: %+v\n}",
		reflect.TypeOf(s.SourceID).Name(), reflect.TypeOf(s.TargetID).Name(), reflect.TypeOf(s.M).Name(), s.SourceID, s.TargetID, s.M)
}

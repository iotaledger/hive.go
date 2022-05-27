package model

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/cockroachdb/errors"

	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/generics/lo"
	"github.com/iotaledger/hive.go/generics/objectstorage"
	"github.com/iotaledger/hive.go/serix"
)

type StorableReference[SourceIDType, TargetIDType any] struct {
	SourceID SourceIDType
	TargetID TargetIDType

	sync.RWMutex
	objectstorage.StorableObjectFlags
}

func NewStorableReference[SourceIDType, TargetIDType any](s SourceIDType, t TargetIDType) (new StorableReference[SourceIDType, TargetIDType]) {
	new = StorableReference[SourceIDType, TargetIDType]{
		SourceID: s,
		TargetID: t,
	}
	new.SetModified()
	new.Persist()

	return new
}

func (s *StorableReference[SourceIDType, TargetIDType]) FromBytes(bytes []byte) (err error) {
	s.Lock()
	defer s.Unlock()

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

func (s *StorableReference[SourceIDType, TargetIDType]) FromObjectStorage(key, _ []byte) (err error) {
	if err = s.FromBytes(key); err != nil {
		err = errors.Errorf("failed to load object from object storage: %w", err)
	}

	return
}

func (s *StorableReference[SourceIDType, TargetIDType]) ObjectStorageKey() (key []byte) {
	return lo.PanicOnErr(s.Bytes())
}

func (s *StorableReference[SourceIDType, TargetIDType]) ObjectStorageValue() (value []byte) {
	return nil
}

func (s StorableReference[SourceIDType, TargetIDType]) KeyPartitions() []int {
	var sourceID SourceIDType
	var targetID TargetIDType

	return []int{len(lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), sourceID))), len(lo.PanicOnErr(serix.DefaultAPI.Encode(context.Background(), targetID)))}
}

func (s *StorableReference[SourceIDType, TargetIDType]) Bytes() (bytes []byte, err error) {
	s.RLock()
	defer s.RUnlock()

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

func (s *StorableReference[SourceIDType, TargetIDType]) String() string {
	return fmt.Sprintf("StorableReference[%s, %s] {\n\tSourceID: %+v\n\tTargetID: %+v\n}",
		reflect.TypeOf(s.SourceID).Name(), reflect.TypeOf(s.TargetID).Name(), s.SourceID, s.TargetID)
}

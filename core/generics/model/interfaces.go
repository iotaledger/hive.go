package model

import "errors"

var (
	// ErrParseBytesFailed is returned if information can not be parsed from a sequence of bytes.
	ErrParseBytesFailed = errors.New("failed to parse bytes")
)

// PtrType is a type constraint that ensures that all the required methods are available.
type PtrType[OuterModelType any, InnerModelType any] interface {
	*OuterModelType

	New(innerModelType *InnerModelType, cacheBytes ...bool)
	Init()
	InnerModel() *InnerModelType
}

// ReferencePtrType is a type constraint that ensures that all the required methods are available.
type ReferencePtrType[OuterModelType, SourceIDType, TargetIDType any] interface {
	*OuterModelType

	New(SourceIDType, TargetIDType)
	Init()
	SourceID() SourceIDType
	TargetID() TargetIDType
}

// ReferenceWithMetadataPtrType is a type constraint that ensures that all the required methods are available.
type ReferenceWithMetadataPtrType[OuterModelType, SourceIDType, TargetIDType, InnerModelType any] interface {
	*OuterModelType

	New(SourceIDType, TargetIDType, *InnerModelType)
	Init()
	SourceID() SourceIDType
	TargetID() TargetIDType
	InnerModel() *InnerModelType
}

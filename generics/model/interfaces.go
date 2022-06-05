package model

// PtrType is a type constraint that ensures that all the required methods are available.
type PtrType[OuterModelType any, InnerModelType any] interface {
	*OuterModelType

	New(*InnerModelType)
	Init()
	InnerModel() *InnerModelType
}

// PtrType is a type constraint that ensures that all the required methods are available.
type ReferencePtrType[OuterModelType, SourceIDType, TargetIDType any] interface {
	*OuterModelType

	InitSource(SourceIDType)
	Source() SourceIDType
}

package model

type outerModelPtr[OuterModelType any, InnerModelType any] interface {
	*OuterModelType

	setM(*InnerModelType)
	m() *InnerModelType
}

type outerStorableModelPtr[OuterModelType any, InnerModelType any] interface {
	*OuterModelType

	Init(*InnerModelType)
	InnerModel() *InnerModelType
}

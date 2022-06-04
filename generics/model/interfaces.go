package model

type outerModelPtr[OuterModelType any, InnerModelType any] interface {
	*OuterModelType

	setM(*InnerModelType)
	m() *InnerModelType
}

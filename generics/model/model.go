package model

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/iotaledger/hive.go/serix"
)

// Model is the base type for all models. It should be embedded in a wrapper type.
// It provides locking and serialization primitives.
type Model[OuterModelType any, OuterModelPtrType modelPtr[OuterModelType, InnerModelType], InnerModelType any] struct {
	M InnerModelType `serix:"0"`

	sync.RWMutex
}

// New returns a new model instance.
func New[OuterModelType, InnerModelType any, OuterModelPtrType modelPtr[OuterModelType, InnerModelType]](model *InnerModelType) (newInstance *OuterModelType) {
	newInstance = new(OuterModelType)
	typedInstance := (OuterModelPtrType)(newInstance)
	typedInstance.Init(model)

	return newInstance
}

func (m *Model[OuterModelType, OuterModelPtrType, InnerModelType]) Init(innerModel *InnerModelType) {
	m.M = *innerModel
}

func (m *Model[OuterModelType, OuterModelPtrType, InnerModelType]) InnerModel() *InnerModelType {
	return &m.M
}

// FromBytes deserializes a model from a byte slice.
func (m *Model[OuterModelType, OuterModelPtrType, InnerModelType]) FromBytes(bytes []byte) (err error) {
	m.Lock()
	defer m.Unlock()

	outerModel := new(OuterModelType)
	_, err = serix.DefaultAPI.Decode(context.Background(), bytes, outerModel, serix.WithValidation())
	m.M = *(OuterModelPtrType)(outerModel).InnerModel()

	return
}

// Bytes serializes a model to a byte slice.
func (m *Model[OuterModelType, OuterModelPtrType, InnerModelType]) Bytes() (bytes []byte, err error) {
	m.RLock()
	defer m.RUnlock()

	outerModel := new(OuterModelType)
	(OuterModelPtrType)(outerModel).Init(&m.M)

	return serix.DefaultAPI.Encode(context.Background(), outerModel, serix.WithValidation())
}

// String returns a string representation of the model.
func (m *Model[OuterModelType, OuterModelPtrType, InnerModelType]) String() string {
	var outerModelType OuterModelType
	return fmt.Sprintf("%s {\n\t%s\n}", reflect.TypeOf(outerModelType).Name(), m.modelString())
}

func (m *Model[OuterModelType, OuterModelPtrType, InnerModelType]) modelString() (humanReadable string) {
	return strings.TrimRight(strings.TrimLeft(fmt.Sprintf("%+v", m.M), "{"), "}")
}

// modelPtr is a type constraint that ensures that all the required methods are available in the OuterModelType.
type modelPtr[OuterModelType any, InnerModelType any] interface {
	*OuterModelType

	Init(*InnerModelType)
	InnerModel() *InnerModelType
}

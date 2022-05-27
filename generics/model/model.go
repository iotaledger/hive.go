package model

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/iotaledger/hive.go/serix"
)

type Model[ModelType any] struct {
	M ModelType `serix:"0"`

	sync.RWMutex
}

func New[ModelType any](model ModelType) (new Model[ModelType]) {
	new = Model[ModelType]{
		M: model,
	}

	return new
}

func (m *Model[ModelType]) FromBytes(bytes []byte) (err error) {
	m.Lock()
	defer m.Unlock()

	_, err = serix.DefaultAPI.Decode(context.Background(), bytes, &m.M, serix.WithValidation())
	return
}

func (m *Model[ModelType]) Bytes() (bytes []byte, err error) {
	m.RLock()
	defer m.RUnlock()

	return serix.DefaultAPI.Encode(context.Background(), m.M, serix.WithValidation())
}

func (m *Model[ModelType]) String() string {
	return fmt.Sprintf("Model[%s] %+v", reflect.TypeOf(m.M).Name(), m.M)
}

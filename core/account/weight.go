package account

import (
	"context"

	"github.com/iotaledger/hive.go/core/index"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

// Weight is a weight annotated with the slot it was last updated in.
type Weight[I index.Type] struct {
	Value      int64 `serix:"0"`
	UpdateTime I     `serix:"1"`
}

// NewWeight creates a new Weight instance.
func NewWeight[I index.Type](value int64, updateTime I) *Weight[I] {
	return &Weight[I]{
		Value:      value,
		UpdateTime: updateTime,
	}
}

// Bytes returns a serialized version of the Weight.
func (w Weight[I]) Bytes() ([]byte, error) {
	return serix.DefaultAPI.Encode(context.Background(), w)
}

// FromBytes parses a serialized version of the Weight.
func (w *Weight[I]) FromBytes(bytes []byte) (int, error) {
	return serix.DefaultAPI.Decode(context.Background(), bytes, w)
}

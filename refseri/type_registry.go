package refseri

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// TypeRegistry stores mapping between typeID that is serialized and actual implementation type.
type TypeRegistry struct {
	registryLock sync.RWMutex

	typeID       uint32
	typeIDToType map[uint32]reflect.Type
	typeToTypeID map[reflect.Type]uint32
}

// ErrTypeNotRegistered error returned when trying to encode/decode a type that was not registered
var ErrTypeNotRegistered = errors.New("type not registered")

// ErrAlreadyRegistered error returned when trying to register a type multiple times
var ErrAlreadyRegistered = errors.New("type already registered")

// NewTypeRegistry creates a new type registry
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		typeID:       0,
		typeIDToType: map[uint32]reflect.Type{},
		typeToTypeID: map[reflect.Type]uint32{},
	}
}

// RegisterType registers new type mapping by assigning incremented numeric value.
func (r *TypeRegistry) RegisterType(value interface{}) error {
	r.registryLock.Lock()
	defer r.registryLock.Unlock()

	valueType := reflect.TypeOf(value)
	if _, exists := r.typeToTypeID[valueType]; exists {
		return fmt.Errorf("%w: type %v", ErrAlreadyRegistered, valueType)
	}
	r.typeIDToType[r.typeID] = valueType
	r.typeToTypeID[valueType] = r.typeID
	r.typeID++
	return nil
}

// EncodeType returns numeric value registered for type t. Returns error if type is not registered.
func (r *TypeRegistry) EncodeType(t reflect.Type) (typeID uint32, err error) {
	r.registryLock.RLock()
	defer r.registryLock.RUnlock()
	typeID, exists := r.typeToTypeID[t]
	if !exists {
		err = fmt.Errorf("%w: type %v", ErrTypeNotRegistered, t)
		return
	}
	return
}

// DecodeType returns type registered for identifier t. Returns error if identifier is not known.
func (r *TypeRegistry) DecodeType(t uint32) (reflect.Type, error) {
	r.registryLock.RLock()
	defer r.registryLock.RUnlock()
	mappedType, exists := r.typeIDToType[t]
	if !exists {
		return nil, fmt.Errorf("%w: typeID %v", ErrTypeNotRegistered, t)
	}
	return mappedType, nil
}

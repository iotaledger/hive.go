package autoserializer

import (
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
		return fmt.Errorf("type %v has already been registered", valueType)
	}
	r.typeIDToType[r.typeID] = valueType
	r.typeToTypeID[valueType] = r.typeID
	r.typeID++
	return nil
}

// EncodeType returns numeric value registered for type t. Returns error if type is not registered.
func (r *TypeRegistry) EncodeType(t reflect.Type) (uint32, error) {
	r.registryLock.RLock()
	defer r.registryLock.RUnlock()
	typeID, exists := r.typeToTypeID[t]
	if !exists {
		return 0, fmt.Errorf("type %v is not registered", t)
	}
	return typeID, nil
}

// DecodeType returns type registered for identifier t. Returns error if identifier is not known.
func (r *TypeRegistry) DecodeType(t uint32) (reflect.Type, error) {
	r.registryLock.RLock()
	defer r.registryLock.RUnlock()
	mappedType, exists := r.typeIDToType[t]
	if !exists {
		return nil, fmt.Errorf("typeID %v is not registered", t)
	}
	return mappedType, nil
}

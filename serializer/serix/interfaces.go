package serix

import (
	"reflect"
	"sync"

	hiveorderedmap "github.com/iotaledger/hive.go/ds/orderedmap"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	ErrInterfaceUnderlyingTypeNotRegistered = ierrors.New("underlying type hasn't been registered for interface type")
)

// InterfaceObjects holds all the information about the objects
// that are registered to the same interface.
type InterfaceObjects struct {
	typeDenotation serializer.TypeDenotationType
	fromCodeToType *hiveorderedmap.OrderedMap[uint32, reflect.Type]
	fromTypeToCode *hiveorderedmap.OrderedMap[reflect.Type, uint32]
}

func NewInterfaceObjects(typeDenotation serializer.TypeDenotationType) *InterfaceObjects {
	return &InterfaceObjects{
		typeDenotation: typeDenotation,
		fromCodeToType: hiveorderedmap.New[uint32, reflect.Type](),
		fromTypeToCode: hiveorderedmap.New[reflect.Type, uint32](),
	}
}

func (i *InterfaceObjects) TypeDenotation() serializer.TypeDenotationType {
	return i.typeDenotation
}

func (i *InterfaceObjects) AddObject(objCode uint32, objType reflect.Type) {
	i.fromCodeToType.Set(objCode, objType)
	i.fromTypeToCode.Set(objType, objCode)
}

func (i *InterfaceObjects) HasObjectType(objType reflect.Type) bool {
	_, exists := i.fromTypeToCode.Get(objType)

	return exists
}

func (i *InterfaceObjects) GetObjectTypeByCode(objCode uint32) (reflect.Type, bool) {
	objType, exists := i.fromCodeToType.Get(objCode)

	return objType, exists
}

func (i *InterfaceObjects) GetObjectCodeByType(objType reflect.Type) (uint32, bool) {
	objCode, exists := i.fromTypeToCode.Get(objType)

	return objCode, exists
}

func (i *InterfaceObjects) ForEachObjectCode(f func(objCode uint32, objType reflect.Type) bool) {
	i.fromTypeToCode.ForEach(func(objType reflect.Type, objCode uint32) bool {
		return f(objCode, objType)
	})
}

func (i *InterfaceObjects) ForEachObjectType(f func(objType reflect.Type, objCode uint32) bool) {
	i.fromCodeToType.ForEach(func(objCode uint32, objType reflect.Type) bool {
		return f(objType, objCode)
	})
}

type InterfacesRegistry struct {
	// the registered interfaces and their known objects
	registryMutex sync.RWMutex
	registry      *hiveorderedmap.OrderedMap[reflect.Type, *InterfaceObjects]
}

func NewInterfacesRegistry() *InterfacesRegistry {
	return &InterfacesRegistry{
		registry: hiveorderedmap.New[reflect.Type, *InterfaceObjects](),
	}
}

func (r *InterfacesRegistry) Has(objType reflect.Type) bool {
	_, exists := r.Get(objType)

	return exists
}

func (r *InterfacesRegistry) Get(objType reflect.Type) (*InterfaceObjects, bool) {
	r.registryMutex.RLock()
	defer r.registryMutex.RUnlock()

	return r.registry.Get(objType)
}

func (r *InterfacesRegistry) ForEach(consumer func(objType reflect.Type, interfaceObjects *InterfaceObjects) bool) {
	r.registryMutex.RLock()
	defer r.registryMutex.RUnlock()

	r.registry.ForEach(func(objType reflect.Type, interfaceObjects *InterfaceObjects) bool {
		return consumer(objType, interfaceObjects)
	})
}

func (r *InterfacesRegistry) RegisterInterfaceObjects(typeSettingsRegistry *TypeSettingsRegistry, iType interface{}, objs ...interface{}) error {
	ptrType := reflect.TypeOf(iType)
	if ptrType == nil {
		return ierrors.New("'iType' is a nil interface, it needs to be a pointer to an interface")
	}

	if ptrType.Kind() != reflect.Ptr {
		return ierrors.Errorf("'iType' parameter must be a pointer, got %s", ptrType.Kind())
	}

	iTypeReflect := ptrType.Elem()
	if iTypeReflect.Kind() != reflect.Interface {
		return ierrors.Errorf(
			"'iType' pointer must contain an interface, got %s", iTypeReflect.Kind())
	}

	if len(objs) == 0 {
		return nil
	}

	r.registryMutex.Lock()
	defer r.registryMutex.Unlock()

	iRegistry, exists := r.registry.Get(iTypeReflect)
	if !exists {
		// get the object metadata for the first object
		objMeta, err := typeSettingsRegistry.GetObjectMetadata(objs[0])
		if err != nil {
			return err
		}

		iRegistry = NewInterfaceObjects(objMeta.TypeDenotation)
	}

	for _, obj := range objs {
		objMeta, err := typeSettingsRegistry.GetObjectMetadata(obj)
		if err != nil {
			return err
		}

		if iRegistry.TypeDenotation() != objMeta.TypeDenotation {
			firstObj := objs[0]

			return ierrors.Errorf(
				"all registered objects must have the same type denotation: object %T has %s and object %T has %s",
				firstObj, iRegistry.TypeDenotation(), obj, objMeta.TypeDenotation,
			)
		}

		iRegistry.AddObject(objMeta.Code, objMeta.Type)
	}

	if !exists {
		r.registry.Set(iTypeReflect, iRegistry)
	}

	return nil
}

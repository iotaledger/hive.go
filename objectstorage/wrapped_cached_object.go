package objectstorage

import (
	"fmt"
	"strconv"

	"github.com/iotaledger/hive.go/reflect"
)

type WrappedCachedObject interface {
	CachedObject

	SetIdentifier(identifier string)
	GetIdentifier() string
}

type WrappedCachedObjectImpl struct {
	*CachedObjectImpl

	identifier string
}

func (wrappedCachedObject *WrappedCachedObjectImpl) Retain() CachedObject {
	baseCachedObject := wrappedCachedObject.CachedObjectImpl
	baseCachedObject.Retain()

	caller := reflect.GetCaller(1)

	result := baseCachedObject.objectStorage.optionalWrap(baseCachedObject, 1)
	result.SetIdentifier(caller.File + ":" + strconv.Itoa(caller.Line))

	fmt.Println(caller.File + ":" + strconv.Itoa(caller.Line))

	// register identifier in debug list

	return result
}

func (wrappedCachedObject *WrappedCachedObjectImpl) Release() {
	baseCachedObject := wrappedCachedObject.CachedObjectImpl

	// unregister identifier in debug list

	baseCachedObject.Release()
}

func (wrappedCachedObject *WrappedCachedObjectImpl) SetIdentifier(identifier string) {
	wrappedCachedObject.identifier = identifier
}

func (wrappedCachedObject *WrappedCachedObjectImpl) GetIdentifier() string {
	return wrappedCachedObject.identifier
}

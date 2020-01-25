package objectstorage

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

	return baseCachedObject.objectStorage.optionalWrap(baseCachedObject, 2)
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

package objectstorage

type CachedObjects []*CachedObjectImpl

func (cachedObjects CachedObjects) Release() {
	for _, cachedObject := range cachedObjects {
		cachedObject.Release()
	}
}

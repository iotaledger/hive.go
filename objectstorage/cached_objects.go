package objectstorage

type CachedObjects []CachedObject

func (cachedObjects CachedObjects) Release() {
	for _, cachedObject := range cachedObjects {
		cachedObject.Release()
	}
}

package objectstorage

type CachedObjects []CachedObject

func (cachedObjects CachedObjects) Release(force ...bool) {
	for _, cachedObject := range cachedObjects {
		cachedObject.Release(force...)
	}
}

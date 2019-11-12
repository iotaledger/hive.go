package objectstorage

type CachedObjects []*CachedObject

func (cachedObjects CachedObjects) Release() {
	for _, cachedObject := range cachedObjects {
		cachedObject.Release()
	}
}

func (cachedObjects CachedObjects) Store() {
	for _, cachedObject := range cachedObjects {
		cachedObject.Store()
	}
}

package objectstorage

import (
	"sync/atomic"
)

type StorableObjectFlags struct {
	persist  atomic.Bool
	delete   atomic.Bool
	modified atomic.Bool
}

func (of *StorableObjectFlags) SetModified(modified ...bool) (wasSet bool) {
	return of.modified.Swap(len(modified) == 0 || modified[0])
}

func (of *StorableObjectFlags) IsModified() bool {
	return of.modified.Load()
}

//nolint:predeclared // lets keep this for now
func (of *StorableObjectFlags) Delete(delete ...bool) (wasSet bool) {
	wasSet = of.delete.Swap(len(delete) == 0 || delete[0])
	of.modified.Store(true)

	return wasSet
}

func (of *StorableObjectFlags) IsDeleted() bool {
	return of.delete.Load()
}

func (of *StorableObjectFlags) Persist(persist ...bool) (wasSet bool) {
	if len(persist) == 0 || persist[0] {
		wasSet = of.persist.Swap(true)
		of.delete.Store(false)
	} else {
		wasSet = of.persist.Swap(false)
	}

	return wasSet
}

// ShouldPersist returns "true" if this object is going to be persisted.
func (of *StorableObjectFlags) ShouldPersist() bool {
	return of.persist.Load()
}

package objectstorage

import (
	"go.uber.org/atomic"
)

type StorableObjectFlags struct {
	persist  atomic.Bool
	delete   atomic.Bool
	modified atomic.Bool
}

func (of *StorableObjectFlags) SetModified(modified bool) (wasSet bool) {
	return of.modified.Swap(modified)
}

func (of *StorableObjectFlags) IsModified() bool {
	return of.modified.Load()
}

func (of *StorableObjectFlags) Delete(delete bool) (wasSet bool) {
	wasSet = of.delete.Swap(delete)
	of.modified.Store(true)
	return wasSet
}

func (of *StorableObjectFlags) IsDeleted() bool {
	return of.delete.Load()
}

func (of *StorableObjectFlags) Persist(persist bool) (wasSet bool) {
	if persist {
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

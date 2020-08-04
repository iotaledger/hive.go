package objectstorage

import (
	"github.com/iotaledger/hive.go/typeutils"
)

type StorableObjectFlags struct {
	persist  typeutils.AtomicBool
	delete   typeutils.AtomicBool
	modified typeutils.AtomicBool
}

func (testObject *StorableObjectFlags) SetModified(modified ...bool) {
	if len(modified) >= 1 {
		testObject.modified.SetTo(modified[0])
	} else {
		testObject.modified.Set()
	}
}

func (testObject *StorableObjectFlags) IsModified() bool {
	return testObject.modified.IsSet()
}

func (testObject *StorableObjectFlags) Delete(delete ...bool) {
	if len(delete) >= 1 {
		testObject.delete.SetTo(delete[0])
	} else {
		testObject.delete.Set()
	}

	testObject.modified.Set()
}

func (testObject *StorableObjectFlags) IsDeleted() bool {
	return testObject.delete.IsSet()
}

func (testObject *StorableObjectFlags) Persist(persist ...bool) {
	if len(persist) >= 1 {
		if persist[0] {
			testObject.persist.Set()
			testObject.delete.UnSet()
		} else {
			testObject.persist.UnSet()
		}
	} else {
		testObject.persist.Set()
		testObject.delete.UnSet()
	}
}

// ShouldPersist returns "true" if this object is going to be persisted.
func (testObject *StorableObjectFlags) ShouldPersist() bool {
	return testObject.persist.IsSet()
}

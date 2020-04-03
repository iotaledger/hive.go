package objectstorage

import (
	"github.com/dgraph-io/badger/v2"
)

// Factory is a utility that offers an api for a more compact creation of multiple ObjectStorage instances from within
// the same package. It will automatically build the corresponding storageId and provide the shared badger instance to
// the created ObjectStorage instances.
type Factory struct {
	badgerInstance *badger.DB
	packagePrefix  byte
}

// NewFactory creates a new Factory with the given ObjectStorage parameters.
func NewFactory(badgerInstance *badger.DB, packagePrefix byte) *Factory {
	return &Factory{
		badgerInstance: badgerInstance,
		packagePrefix:  packagePrefix,
	}
}

// New creates a new ObjectStorage with the given parameters. It combines the storage specific prefix with the package
// prefix, to create a unique storageId for the ObjectStorage.
func (factory *Factory) New(storagePrefix byte, objectFactory StorableObjectFromKey, optionalOptions ...Option) *ObjectStorage {
	return New(factory.badgerInstance, []byte{factory.packagePrefix, storagePrefix}, objectFactory, optionalOptions...)
}

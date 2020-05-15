package objectstorage

import "github.com/iotaledger/hive.go/kvstore"

// Factory is a utility that offers an api for a more compact creation of multiple ObjectStorage instances from within
// the same package. It will automatically build the corresponding storageId and provide the shared badger instance to
// the created ObjectStorage instances.
type Factory struct {
	store         kvstore.KVStore
	packagePrefix byte
}

// NewFactory creates a new Factory with the given ObjectStorage parameters.
func NewFactory(store kvstore.KVStore, packagePrefix byte) *Factory {
	return &Factory{
		store:         store,
		packagePrefix: packagePrefix,
	}
}

// New creates a new ObjectStorage with the given parameters. It combines the store specific prefix with the package
// prefix, to create a unique storageId for the ObjectStorage.
func (factory *Factory) New(storagePrefix byte, objectFactory StorableObjectFromKey, optionalOptions ...Option) *ObjectStorage {
	return New(factory.store.WithRealm([]byte{factory.packagePrefix, storagePrefix}), objectFactory, optionalOptions...)
}

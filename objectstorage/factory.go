package objectstorage

// Factory is a utility that offers an api for a more compact creation of multiple ObjectStorage instances from within
// the same package. It will automatically build the corresponding storageId and provide the shared badger instance to
// the created ObjectStorage instances.
type Factory struct {
	storage       Storage
	packagePrefix byte
}

// NewFactory creates a new Factory with the given ObjectStorage parameters.
func NewFactory(storage Storage, packagePrefix byte) *Factory {
	return &Factory{
		storage:       storage,
		packagePrefix: packagePrefix,
	}
}

// New creates a new ObjectStorage with the given parameters. It combines the storage specific prefix with the package
// prefix, to create a unique storageId for the ObjectStorage.
func (factory *Factory) New(storagePrefix byte, objectFactory StorableObjectFromKey, optionalOptions ...Option) *ObjectStorage {
	return New(factory.storage.WithRealm([]byte{factory.packagePrefix, storagePrefix}), objectFactory, optionalOptions...)
}

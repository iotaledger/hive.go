package module

// Provider is a function that returns a Module.
type Provider[ContainerType any, ModuleType Module] func(ContainerType) ModuleType

// Provide turns a constructor into a provider.
func Provide[ContainerType any, ModuleType Module](constructor func(ContainerType) ModuleType) Provider[ContainerType, ModuleType] {
	return func(c ContainerType) ModuleType {
		return constructor(c)
	}
}

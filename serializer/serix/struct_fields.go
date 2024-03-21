package serix

import (
	"reflect"
	"sort"
	"sync"

	"github.com/iotaledger/hive.go/ierrors"
)

type structField struct {
	name         string
	isUnexported bool
	index        int
	fType        reflect.Type
	isEmbedded   bool
	settings     TagSettings
}

// parseStructFields parses the struct fields of the given struct type.
// It returns a slice of structField (only the ones with valid serix tags).
// Neither the interfacesRegistry, the typeSettingsRegistry nor the validatorsRegistry are used by this function.
// The returned result is only based on the struct tags.
func parseStructFields(structType reflect.Type) ([]structField, error) {
	structFields := make([]structField, 0, structType.NumField())

	serixPosition := 0
	for i := range structType.NumField() {
		field := structType.Field(i)

		isUnexported := field.PkgPath != ""
		isEmbedded := field.Anonymous
		isStruct := isUnderlyingStruct(field.Type)
		isInterface := isUnderlyingInterface(field.Type)
		isEmbeddedStruct := isEmbedded && isStruct
		isEmbeddedInterface := isEmbedded && isInterface

		if isUnexported && !isEmbeddedStruct && !isEmbeddedInterface {
			continue
		}

		tag, ok := field.Tag.Lookup("serix")
		if !ok {
			continue
		}

		tSettings, err := ParseSerixSettings(tag, serixPosition)
		if err != nil {
			return nil, ierrors.Wrapf(err, "failed to parse serix struct tag for field %s", field.Name)
		}
		serixPosition++

		if tSettings.isOptional {
			if field.Type.Kind() != reflect.Ptr && field.Type.Kind() != reflect.Interface {
				return nil, ierrors.Errorf(
					"struct field %s is invalid: "+
						"'optional' setting can only be used with pointers or interfaces, got %s",
					field.Name, field.Type.Kind())
			}

			if isEmbeddedStruct {
				return nil, ierrors.Errorf(
					"struct field %s is invalid: 'optional' setting can't be used with embedded structs",
					field.Name)
			}

			if isEmbeddedInterface {
				return nil, ierrors.Errorf(
					"struct field %s is invalid: 'optional' setting can't be used with embedded interfaces",
					field.Name)
			}
		}

		if tSettings.inlined && isUnexported {
			return nil, ierrors.Errorf(
				"struct field %s is invalid: 'inlined' setting can't be used with unexported types",
				field.Name)
		}

		if !tSettings.inlined && isEmbeddedInterface {
			return nil, ierrors.Errorf(
				"struct field %s is invalid: 'inlined' setting needs to be used for embedded interfaces",
				field.Name)
		}

		structFields = append(structFields, structField{
			name:         field.Name,
			isUnexported: isUnexported,
			index:        i,
			fType:        field.Type,
			isEmbedded:   isEmbeddedStruct || isEmbeddedInterface,
			settings:     tSettings,
		})
	}
	sort.Slice(structFields, func(i, j int) bool {
		return structFields[i].settings.position < structFields[j].settings.position
	})

	return structFields, nil
}

type structFieldsCache struct {
	cacheMutex sync.RWMutex
	cache      map[reflect.Type][]structField
}

func newStructFieldsCache() *structFieldsCache {
	return &structFieldsCache{
		cache: make(map[reflect.Type][]structField),
	}
}

func (c *structFieldsCache) Get(structType reflect.Type) ([]structField, bool) {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()
	structFields, exists := c.cache[structType]

	return structFields, exists
}

func (c *structFieldsCache) Set(structType reflect.Type, structFields []structField) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.cache[structType] = structFields
}

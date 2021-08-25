package autoserializer

import (
	"fmt"
	"go/types"
	"reflect"
	"strconv"
	"sync"
	"unicode"
)

type fieldCache struct {
	lock             sync.Mutex
	structFieldCache map[reflect.Type][]serializationMetadata
}

// NewFieldCache creates and returns new fieldCache.
func NewFieldCache() *fieldCache {
	return &fieldCache{
		structFieldCache: make(map[reflect.Type][]serializationMetadata),
	}
}

type serializationMetadata struct {
	idx              int
	unpack           bool
	lengthPrefixType types.BasicKind
	maxSliceLength   int
	minSliceLength   int
}

// GetFields returns struct fields that are available for serialization. It caches the fields so consecutive calls for the same time can use previously extracted values.
func (c *fieldCache) GetFields(structType reflect.Type) ([]serializationMetadata, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.structFieldCache == nil {
		c.structFieldCache = make(map[reflect.Type][]serializationMetadata)
	}
	if cachedFields, ok := c.structFieldCache[structType]; ok {
		return cachedFields, nil
	}
	numFields := structType.NumField()
	cachedFields := make([]serializationMetadata, 0, numFields)
	for i := 0; i < numFields; i++ {
		field := structType.Field(i)
		var sm serializationMetadata
		sm.lengthPrefixType = types.Uint8
		switch field.Tag.Get("serialize") {
		case "unpack":
			sm.unpack = true
			fallthrough
		case "true":
			if !sm.unpack && unicode.IsLower(rune(field.Name[0])) {
				return nil, fmt.Errorf("can't marshal un-exported field '%s'", structType.Field(i).Name)
			}
			sm.idx = i

			if v := field.Tag.Get("minLen"); v != "" {
				minLen, err := strconv.Atoi(v)
				if err != nil {
					return nil, err
				}
				sm.minSliceLength = minLen
			}

			if v := field.Tag.Get("maxLen"); v != "" {
				maxLen, err := strconv.Atoi(v)
				if err != nil {
					return nil, err
				}
				sm.maxSliceLength = maxLen
			}

			if v := field.Tag.Get("lenPrefixBytes"); v != "" {
				prefixBytes, err := strconv.Atoi(v)
				if err != nil {
					return nil, err
				}
				switch prefixBytes {
				case 1:
					sm.lengthPrefixType = types.Uint8
				case 2:
					sm.lengthPrefixType = types.Uint16
				case 4:
					sm.lengthPrefixType = types.Uint32
				}
			}
			cachedFields = append(cachedFields, sm)

		}
	}
	c.structFieldCache[structType] = cachedFields
	return cachedFields, nil
}

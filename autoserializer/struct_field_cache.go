package autoserializer

import (
	"fmt"
	"reflect"
	"sync"
	"unicode"
)

type fieldCache struct {
	lock sync.Mutex

	structFieldCache map[reflect.Type][]int
}

func NewFieldCache() *fieldCache {
	return &fieldCache{
		structFieldCache: make(map[reflect.Type][]int),
	}
}

func (c *fieldCache) GetFields(structType reflect.Type) ([]int, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.structFieldCache == nil {
		c.structFieldCache = make(map[reflect.Type][]int)
	}
	if cachedFields, ok := c.structFieldCache[structType]; ok {
		return cachedFields, nil
	}
	numFields := structType.NumField()
	cachedFields := make([]int, 0, numFields)
	for i := 0; i < numFields; i++ {
		field := structType.Field(i)
		if field.Tag.Get("serialize") == "true" {
			if unicode.IsLower(rune(field.Name[0])) {
				return nil, fmt.Errorf("can't marshal un-exported field '%s'", structType.Field(i).Name)
			}
			cachedFields = append(cachedFields, i)
		}
	}
	c.structFieldCache[structType] = cachedFields
	return cachedFields, nil
}

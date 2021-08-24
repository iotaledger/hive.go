package autoserializer

import (
	"reflect"
	"sync"
)

type fieldCache struct {
	lock sync.Mutex

	structFieldCache map[reflect.Type][]int
}

// NewFieldCache creates and returns new fieldCache.
func NewFieldCache() *fieldCache {
	return &fieldCache{
		structFieldCache: make(map[reflect.Type][]int),
	}
}

// GetFields returns struct fields that are available for serialization. It caches the fields so consecutive calls for the same time can use previously extracted values.
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
		switch field.Tag.Get("serialize") {
		case "true":
			fallthrough
		case "unpack":
			/*
				if unicode.IsLower(rune(field.Name[0])) {
					return nil, fmt.Errorf("can't marshal un-exported field '%s'", structType.Field(i).Name)
				}
			*/
			cachedFields = append(cachedFields, i)
		}
	}
	c.structFieldCache[structType] = cachedFields
	return cachedFields, nil
}

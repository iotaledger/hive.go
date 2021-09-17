package orderedmap

import (
	"reflect"
	"sync"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/reflectionserializer"
)

// OrderedMap provides a concurrent-safe ordered map.
type OrderedMap struct {
	head       *Element
	tail       *Element
	dictionary map[interface{}]*Element
	size       int
	mutex      sync.RWMutex
}

// New returns a new *OrderedMap.
func New() *OrderedMap {
	return &OrderedMap{
		dictionary: make(map[interface{}]*Element),
	}
}

// Head returns the first map entry.
func (orderedMap *OrderedMap) Head() (key, value interface{}, exists bool) {
	orderedMap.mutex.RLock()
	defer orderedMap.mutex.RUnlock()

	if exists = orderedMap.head != nil; !exists {
		return
	}
	key = orderedMap.head.key
	value = orderedMap.head.value

	return
}

// Tail returns the last map entry.
func (orderedMap *OrderedMap) Tail() (key, value interface{}, exists bool) {
	orderedMap.mutex.RLock()
	defer orderedMap.mutex.RUnlock()

	if exists = orderedMap.tail != nil; !exists {
		return
	}
	key = orderedMap.tail.key
	value = orderedMap.tail.value

	return
}

// Has returns if an entry with the given key exists.
func (orderedMap *OrderedMap) Has(key interface{}) (has bool) {
	orderedMap.mutex.RLock()
	defer orderedMap.mutex.RUnlock()

	_, has = orderedMap.dictionary[key]

	return
}

// Get returns the value mapped to the given key if exists.
func (orderedMap *OrderedMap) Get(key interface{}) (interface{}, bool) {
	orderedMap.mutex.RLock()
	defer orderedMap.mutex.RUnlock()

	orderedMapElement, orderedMapElementExists := orderedMap.dictionary[key]
	if !orderedMapElementExists {
		return nil, false
	}
	return orderedMapElement.value, true
}

// Set adds a key-value pair to the orderedMap. It returns false if the same pair already exists.
func (orderedMap *OrderedMap) Set(key interface{}, newValue interface{}) bool {
	if oldValue, oldValueExists := orderedMap.Get(key); oldValueExists && oldValue == newValue {
		return false
	}

	orderedMap.mutex.Lock()
	defer orderedMap.mutex.Unlock()

	if oldValue, oldValueExists := orderedMap.dictionary[key]; oldValueExists {
		if oldValue.value == newValue {
			return false
		}

		oldValue.value = newValue

		return true
	}

	newElement := &Element{
		key:   key,
		value: newValue,
	}

	if orderedMap.head == nil {
		orderedMap.head = newElement
	} else {
		orderedMap.tail.next = newElement
		newElement.prev = orderedMap.tail
	}
	orderedMap.tail = newElement
	orderedMap.size++

	orderedMap.dictionary[key] = newElement

	return true
}

// ForEach iterates through the orderedMap and calls the consumer function for every element.
// The iteration can be aborted by returning false in the consumer.
func (orderedMap *OrderedMap) ForEach(consumer func(key, value interface{}) bool) bool {
	orderedMap.mutex.RLock()
	currentEntry := orderedMap.head
	orderedMap.mutex.RUnlock()

	for currentEntry != nil {
		if !consumer(currentEntry.key, currentEntry.value) {
			return false
		}

		orderedMap.mutex.RLock()
		currentEntry = currentEntry.next
		orderedMap.mutex.RUnlock()
	}

	return true
}

// ForEachReverse iterates through the orderedMap in reverse order and calls the consumer function for every element.
// The iteration can be aborted by returning false in the consumer.
func (orderedMap *OrderedMap) ForEachReverse(consumer func(key, value interface{}) bool) bool {
	orderedMap.mutex.RLock()
	currentEntry := orderedMap.tail
	orderedMap.mutex.RUnlock()

	for currentEntry != nil {
		if !consumer(currentEntry.key, currentEntry.value) {
			return false
		}

		orderedMap.mutex.RLock()
		currentEntry = currentEntry.prev
		orderedMap.mutex.RUnlock()
	}

	return true
}

// Clear removes all elements from the OrderedMap.
func (orderedMap *OrderedMap) Clear() {
	orderedMap.mutex.Lock()
	defer orderedMap.mutex.Unlock()

	orderedMap.head = nil
	orderedMap.tail = nil
	orderedMap.size = 0
	orderedMap.dictionary = make(map[interface{}]*Element)
}

// Delete deletes the given key (and related value) from the orderedMap.
// It returns false if the key is not found.
func (orderedMap *OrderedMap) Delete(key interface{}) bool {
	if _, valueExists := orderedMap.Get(key); !valueExists {
		return false
	}

	orderedMap.mutex.Lock()
	defer orderedMap.mutex.Unlock()

	value, valueExists := orderedMap.dictionary[key]
	if !valueExists {
		return false
	}

	delete(orderedMap.dictionary, key)
	orderedMap.size--

	if value.prev != nil {
		value.prev.next = value.next
	} else {
		orderedMap.head = value.next
	}

	if value.next != nil {
		value.next.prev = value.prev
	} else {
		orderedMap.tail = value.prev
	}

	return true
}

// Size returns the size of the orderedMap.
func (orderedMap *OrderedMap) Size() int {
	orderedMap.mutex.RLock()
	defer orderedMap.mutex.RUnlock()

	return orderedMap.size
}

// SerializeBytes implements the reflectionserializer.BinarySerializer interface for serialization.
func (orderedMap *OrderedMap) SerializeBytes(m *reflectionserializer.SerializationManager, fieldMetadata reflectionserializer.FieldMetadata) (data []byte, err error) {
	buffer := marshalutil.New()
	err = reflectionserializer.WriteLen(orderedMap.Size(), fieldMetadata.LengthPrefixType, buffer)
	if err != nil {
		return
	}
	err = reflectionserializer.ValidateLength(orderedMap.Size(), fieldMetadata.MinSliceLength, fieldMetadata.MaxSliceLength)
	if err != nil {
		return nil, err
	}
	if orderedMap.Size() == 0 {
		return buffer.Bytes(), nil
	}
	var encodedKeyType uint32
	var encodedValType uint32
	orderedMap.ForEach(func(key, value interface{}) bool {
		encodedKeyType, err = m.EncodeType(reflect.TypeOf(key))
		if err != nil {
			return false
		}
		encodedValType, err = m.EncodeType(reflect.TypeOf(value))
		if err != nil {
			return false
		}
		buffer.WriteUint32(encodedKeyType)
		buffer.WriteUint32(encodedValType)
		err = m.SerializeValue(reflect.ValueOf(key), fieldMetadata, buffer)
		if err != nil {
			return false
		}
		err = m.SerializeValue(reflect.ValueOf(value), fieldMetadata, buffer)
		return err == nil
	})
	if err != nil {
		return
	}
	return buffer.Bytes(), nil
}

// DeserializeBytes implements the reflectionserializer.BinaryDeserializer interface for deserialization.
func (orderedMap *OrderedMap) DeserializeBytes(buffer *marshalutil.MarshalUtil, m *reflectionserializer.SerializationManager, fieldMetadata reflectionserializer.FieldMetadata) (err error) {
	orderedMap.dictionary = make(map[interface{}]*Element)
	var orderedMapSize int
	orderedMapSize, err = reflectionserializer.ReadLen(fieldMetadata.LengthPrefixType, buffer)
	if err != nil {
		return
	}
	err = reflectionserializer.ValidateLength(orderedMapSize, fieldMetadata.MinSliceLength, fieldMetadata.MaxSliceLength)
	if err != nil {
		return err
	}

	if orderedMapSize == 0 {
		return nil
	}

	var encodedKeyType, encodedValueType uint32
	var keyType, valueType reflect.Type

	for i := 0; i < orderedMapSize; i++ {
		encodedKeyType, err = buffer.ReadUint32()
		if err != nil {
			return
		}
		encodedValueType, err = buffer.ReadUint32()
		if err != nil {
			return
		}
		keyType, err = m.DecodeType(encodedKeyType)
		if err != nil {
			return
		}
		valueType, err = m.DecodeType(encodedValueType)
		if err != nil {
			return
		}

		var key, value interface{}
		key, err = m.DeserializeType(keyType, fieldMetadata, buffer)
		if err != nil {
			return
		}
		value, err = m.DeserializeType(valueType, fieldMetadata, buffer)
		if err != nil {
			return
		}
		orderedMap.Set(key, value)
	}
	return nil
}

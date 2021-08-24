package autoserializer

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"time"

	"github.com/iotaledger/hive.go/datastructure/orderedmap"
	"github.com/iotaledger/hive.go/marshalutil"
)

// SerializationManager stores TypeRegistry and is used to automatically serialize and deserialize structures.
type SerializationManager struct {
	*TypeRegistry
	*fieldCache
}

// NewSerializationManager creates new serialization manager with default TypeRegistry.
func NewSerializationManager() *SerializationManager {
	sm := &SerializationManager{
		TypeRegistry: NewTypeRegistry(),
		fieldCache:   NewFieldCache(),
	}

	return sm
}

// Deserialize data into type s and save its value. s must be a pointer type.
func (m *SerializationManager) Deserialize(s interface{}, data []byte) error {
	marshalUtil := marshalutil.New(data)
	value := reflect.ValueOf(s)
	if value.Kind() != reflect.Ptr {
		return errors.New("pointer type is required to perform deserialization")
	}

	result, err := m.deserialize(reflect.TypeOf(s).Elem(), marshalUtil)
	if err != nil {
		return err
	}
	done, err := marshalUtil.DoneReading()
	if err != nil {
		return err
	}
	if !done {
		return errors.New("did not read all bytes from the buffer")
	}
	value.Elem().Set(reflect.ValueOf(result))
	return nil
}

func (m *SerializationManager) deserialize(valueType reflect.Type, marshalUtil *marshalutil.MarshalUtil) (interface{}, error) {
	switch valueType.Kind() {
	case reflect.Bool:
		tmp, err := marshalUtil.ReadBool()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Int8:
		tmp, err := marshalUtil.ReadInt8()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Int16:
		tmp, err := marshalUtil.ReadInt16()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Int32:
		tmp, err := marshalUtil.ReadInt32()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Int64:
		tmp, err := marshalUtil.ReadInt64()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Int:
		tmp, err := marshalUtil.ReadInt64()
		if err != nil {
			return nil, err
		}
		return int(tmp), nil
	case reflect.Uint8:
		tmp, err := marshalUtil.ReadUint8()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Uint16:
		tmp, err := marshalUtil.ReadUint16()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Uint32:
		tmp, err := marshalUtil.ReadUint32()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Uint64:
		tmp, err := marshalUtil.ReadUint64()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Uint:
		tmp, err := marshalUtil.ReadUint64()
		if err != nil {
			return nil, err
		}
		return uint(tmp), nil
	case reflect.Float32:
		tmp, err := marshalUtil.ReadFloat32()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Float64:
		tmp, err := marshalUtil.ReadFloat64()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.String:
		tmp, err := marshalUtil.ReadUint16()
		if err != nil {
			return nil, err
		}
		bytes, err := marshalUtil.ReadBytes(int(tmp))
		if err != nil {
			return nil, err
		}
		restoredString := string(bytes)
		return restoredString, nil
	case reflect.Array:
		restoredArray := reflect.New(valueType).Elem()
		for i := 0; i < valueType.Len(); i++ {
			elem, err := m.deserialize(valueType.Elem(), marshalUtil)
			if err != nil {
				return nil, err
			}
			restoredArray.Index(i).Set(reflect.ValueOf(elem))
		}
		return restoredArray.Interface(), nil
	case reflect.Slice:
		sliceLen, err := marshalUtil.ReadUint8()
		if err != nil {
			return nil, err
		}
		restoredSlice := reflect.New(valueType).Elem()
		if sliceLen == 0 {
			return restoredSlice.Interface(), nil
		}
		for i := 0; i < int(sliceLen); i++ {
			elem, err := m.deserialize(valueType.Elem(), marshalUtil)
			if err != nil {
				return nil, err
			}
			restoredSlice = reflect.Append(restoredSlice, reflect.ValueOf(elem))
		}
		return restoredSlice.Interface(), nil
	case reflect.Map:
		mapSize, err := marshalUtil.ReadUint8()
		if err != nil {
			return nil, err
		}
		restoredMap := reflect.MakeMap(valueType)
		if mapSize == 0 {
			return restoredMap.Interface(), nil
		}
		for i := 0; i < int(mapSize); i++ {
			k, err := m.deserialize(valueType.Key(), marshalUtil)
			if err != nil {
				return nil, err
			}
			v, err := m.deserialize(valueType.Elem(), marshalUtil)
			if err != nil {
				return nil, err
			}
			restoredMap.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
		}
		return restoredMap.Interface(), nil
	case reflect.Ptr:
		nilPointer, err := marshalUtil.ReadByte()
		if err != nil {
			return nil, err
		}
		if nilPointer == 0 {
			p := reflect.New(valueType.Elem())
			return p.Interface(), nil
		} else {
			p := reflect.New(valueType.Elem())
			de, err := m.deserialize(valueType.Elem(), marshalUtil)
			if err != nil {
				return nil, err
			}
			p.Elem().Set(reflect.ValueOf(de))
			return p.Interface(), nil
		}
	case reflect.Struct:
		s, err := m.deserializeStruct(valueType, marshalUtil)
		if err != nil {
			return nil, err
		}
		return s, nil

	case reflect.Interface:
		encodedType, err := marshalUtil.ReadUint32()
		if err != nil {
			return nil, err
		}
		implementationType, err := m.DecodeType(encodedType)
		if err != nil {
			return nil, err
		}
		if !implementationType.Implements(valueType) {
			return nil, fmt.Errorf("couldn't deserialize interface: %s must implement interface %s", implementationType, valueType)
		}
		return m.deserialize(implementationType, marshalUtil)
	}

	return nil, nil
}

func (m *SerializationManager) deserializeStruct(structType reflect.Type, marshalutil *marshalutil.MarshalUtil) (interface{}, error) {

	if structType == reflect.TypeOf(time.Time{}) {
		restoredTime, err := marshalutil.ReadTime()
		if err != nil {
			return nil, err
		}
		return restoredTime, nil
	} else if structType == reflect.TypeOf(orderedmap.OrderedMap{}) {
		restoredMap := orderedmap.New()
		orderedMapSize, err := marshalutil.ReadUint64()
		if err != nil {
			return nil, err
		}
		if orderedMapSize == 0 {
			return *restoredMap, nil
		}
		encodedKeyType, err := marshalutil.ReadUint32()
		if err != nil {
			return nil, err
		}
		encodedValueType, err := marshalutil.ReadUint32()
		if err != nil {
			return nil, err
		}
		keyType, err := m.DecodeType(encodedKeyType)
		if err != nil {
			return nil, err
		}
		valueType, err := m.DecodeType(encodedValueType)
		if err != nil {
			return nil, err
		}

		for i := uint64(0); i < orderedMapSize; i++ {
			key, err := m.deserialize(keyType, marshalutil)
			if err != nil {
				return nil, err
			}
			value, err := m.deserialize(valueType, marshalutil)
			if err != nil {
				return nil, err
			}
			restoredMap.Set(key, value)
		}
		return *restoredMap, nil
	} else {
		restoredStruct := reflect.New(structType).Elem()
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			switch field.Tag.Get("serialize") {
			case "true":
				fieldValue, err := m.deserialize(field.Type, marshalutil)
				if err != nil {
					return nil, err
				}
				restoredStruct.Field(i).Set(reflect.ValueOf(fieldValue).Convert(field.Type))
			case "unpack":
				if !field.Anonymous {
					panic(fmt.Sprintf("unpack on non anonymous field %s", field.Name))
				}

				if field.Type.Kind() != reflect.Struct {
					panic(fmt.Sprintf("unpack on non anonymous struct field %s", field.Name))
				}

				anonEmbeddedStructType := field.Type
				for j := 0; j < anonEmbeddedStructType.NumField(); j++ {
					embStructField := anonEmbeddedStructType.Field(j)
					if embStructField.Tag.Get("serialize") != "true" {
						continue
					}
					embStructFieldVal, err := m.deserialize(embStructField.Type, marshalutil)
					if err != nil {
						return nil, err
					}
					restoredStruct.Field(i).Field(j).Set(reflect.ValueOf(embStructFieldVal).Convert(embStructField.Type))
				}
			}
		}

		return restoredStruct.Interface(), nil
	}
}

// Serialize turns object into bytes.
func (m *SerializationManager) Serialize(s interface{}) ([]byte, error) {
	buffer := marshalutil.New()
	if reflect.TypeOf(s).Kind() == reflect.Ptr {
		if err := m.serialize(reflect.ValueOf(s).Elem(), buffer); err != nil {
			return nil, err
		}
		return buffer.Bytes(), nil
	}
	if err := m.serialize(reflect.ValueOf(s), buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (m *SerializationManager) serialize(value reflect.Value, buffer *marshalutil.MarshalUtil) error {
	var err error
	switch value.Kind() {
	case reflect.Bool:
		buffer.WriteBool(value.Bool())
	case reflect.Int8:
		buffer.WriteInt8(int8(value.Int()))
	case reflect.Int16:
		buffer.WriteInt16(int16(value.Int()))
	case reflect.Int32:
		buffer.WriteInt32(int32(value.Int()))
	case reflect.Int64, reflect.Int:
		buffer.WriteInt64(value.Int())
	case reflect.Uint8:
		buffer.WriteUint8(uint8(value.Uint()))
	case reflect.Uint16:
		buffer.WriteUint16(uint16(value.Uint()))
	case reflect.Uint32:
		buffer.WriteUint32(uint32(value.Uint()))
	case reflect.Uint64, reflect.Uint:
		buffer.WriteUint64(value.Uint())
	case reflect.Float32:
		if math.IsNaN(value.Float()) {
			return fmt.Errorf("NaN float value")
		}
		buffer.WriteFloat32(float32(value.Float()))
	case reflect.Float64:
		if math.IsNaN(value.Float()) {
			return fmt.Errorf("NaN float value")
		}
		buffer.WriteFloat64(value.Float())
	case reflect.String:
		buffer.WriteUint16(uint16(len(value.String())))
		buffer.WriteBytes([]byte(value.String()))
	case reflect.Array:
		for i := 0; i < value.Len(); i++ {
			err = m.serialize(value.Index(i), buffer)
			if err != nil {
				break
			}
		}
	case reflect.Slice:
		// TODO: parametrize this using a struct tag
		buffer.WriteUint8(uint8(value.Len()))

		for i := 0; i < value.Len(); i++ {
			err = m.serialize(value.Index(i), buffer)
			if err != nil {
				break
			}
		}
	case reflect.Map:
		// TODO: parametrize this using a struct tag
		buffer.WriteUint8(uint8(value.Len()))

		keys := value.MapKeys()
		// TODO: conditional sort based on tag
		sort.Slice(keys, keyComparator(keys))
		for _, k := range keys {
			err = m.serialize(k, buffer)
			if err != nil {
				break
			}
			err = m.serialize(value.MapIndex(k), buffer)
			if err != nil {
				break
			}
		}
	case reflect.Ptr:
		if value.IsNil() {
			buffer.WriteByte(byte(0))
		} else {
			buffer.WriteByte(byte(1))
			err = m.serialize(value.Elem(), buffer)
		}
	case reflect.Struct:
		structType := value.Type()
		if structType == reflect.TypeOf(time.Time{}) {
			buffer.WriteTime(value.Interface().(time.Time))
		} else if structType == reflect.TypeOf(orderedmap.OrderedMap{}) {
			err = m.serializeOrderedMap(value, buffer)
			if err != nil {
				break
			}
		} else {
			err = m.serializeStruct(value, buffer)
		}
	case reflect.Interface:
		interfaceType := reflect.TypeOf(value.Interface())
		var encodedType uint32
		encodedType, err = m.EncodeType(interfaceType)
		if err != nil {
			break
		}
		buffer.WriteUint32(encodedType)
		err = m.serialize(value.Elem(), buffer)

	}
	return err
}

func (m *SerializationManager) serializeStruct(value reflect.Value, buffer *marshalutil.MarshalUtil) error {
	structType := value.Type()
	var err error
	serializedFields, err := m.fieldCache.GetFields(structType)
	if err != nil {
		return err
	}
	for _, i := range serializedFields {
		fieldValue := value.Field(i)
		err = m.serialize(fieldValue, buffer)
		if err != nil {
			break
		}
	}
	return err
}

func (m *SerializationManager) serializeOrderedMap(value reflect.Value, buffer *marshalutil.MarshalUtil) error {
	orderedMap := value.Interface().(orderedmap.OrderedMap)
	buffer.WriteUint64(uint64(orderedMap.Size()))
	if orderedMap.Size() == 0 {
		return nil
	}
	k, v, _ := orderedMap.Head()
	encodedKeyType, err := m.EncodeType(reflect.TypeOf(k))
	if err != nil {
		return err
	}
	encodedValType, err := m.EncodeType(reflect.TypeOf(v))
	if err != nil {
		return err
	}
	buffer.WriteUint32(encodedKeyType)
	buffer.WriteUint32(encodedValType)
	orderedMap.ForEach(func(key, value interface{}) bool {
		err = m.serialize(reflect.ValueOf(key), buffer)
		if err != nil {
			return false
		}
		err = m.serialize(reflect.ValueOf(value), buffer)
		return err == nil
	})
	return err
}

func keyComparator(keys []reflect.Value) func(int, int) bool {
	return func(i int, j int) bool {
		a, b := keys[i], keys[j]
		if a.Kind() == reflect.Interface {
			a = a.Elem()
			b = b.Elem()
		}
		switch a.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return a.Int() < b.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return a.Uint() < b.Uint()
		case reflect.Float32, reflect.Float64:
			return a.Float() < b.Float()
		case reflect.String:
			return a.String() < b.String()
		}
		panic("unsupported key type")
	}
}

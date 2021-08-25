package autoserializer

import (
	"errors"
	"fmt"
	"go/types"
	"math"
	"reflect"
	"sort"
	"time"

	"github.com/iotaledger/hive.go/datastructure/orderedmap"
	"github.com/iotaledger/hive.go/marshalutil"
)

const (
	defaultMinLen        = 0
	defaultMaxLen        = 0
	defaultLenPrefixType = types.Uint8
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
	buffer := marshalutil.New(data)
	value := reflect.ValueOf(s)
	if value.Kind() != reflect.Ptr {
		return errors.New("pointer type is required to perform deserialization")
	}

	result, err := m.deserialize(reflect.TypeOf(s).Elem(), defaultMinLen, defaultMaxLen, defaultLenPrefixType, buffer)
	if err != nil {
		return err
	}
	done, err := buffer.DoneReading()
	if err != nil {
		return err
	}
	if !done {
		return errors.New("did not read all bytes from the buffer")
	}
	value.Elem().Set(reflect.ValueOf(result))
	return nil
}

func (m *SerializationManager) deserialize(valueType reflect.Type, minSliceLen, maxSliceLen int, lenPrefixType types.BasicKind, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	switch valueType.Kind() {
	case reflect.Bool:
		tmp, err := buffer.ReadBool()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Int8:
		tmp, err := buffer.ReadInt8()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Int16:
		tmp, err := buffer.ReadInt16()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Int32:
		tmp, err := buffer.ReadInt32()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Int64:
		tmp, err := buffer.ReadInt64()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Int:
		tmp, err := buffer.ReadInt64()
		if err != nil {
			return nil, err
		}
		return int(tmp), nil
	case reflect.Uint8:
		tmp, err := buffer.ReadUint8()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Uint16:
		tmp, err := buffer.ReadUint16()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Uint32:
		tmp, err := buffer.ReadUint32()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Uint64:
		tmp, err := buffer.ReadUint64()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Uint:
		tmp, err := buffer.ReadUint64()
		if err != nil {
			return nil, err
		}
		return uint(tmp), nil
	case reflect.Float32:
		tmp, err := buffer.ReadFloat32()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.Float64:
		tmp, err := buffer.ReadFloat64()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	case reflect.String:
		tmp, err := buffer.ReadUint16()
		if err != nil {
			return nil, err
		}
		bytes, err := buffer.ReadBytes(int(tmp))
		if err != nil {
			return nil, err
		}
		restoredString := string(bytes)
		return restoredString, nil
	case reflect.Array:
		restoredArray := reflect.New(valueType).Elem()
		for i := 0; i < valueType.Len(); i++ {
			elem, err := m.deserialize(valueType.Elem(), minSliceLen, maxSliceLen, lenPrefixType, buffer)
			if err != nil {
				return nil, err
			}
			restoredArray.Index(i).Set(reflect.ValueOf(elem))
		}
		return restoredArray.Interface(), nil
	case reflect.Slice:
		sliceLen, err := readLen(lenPrefixType, buffer)
		if err != nil {
			return nil, err
		}
		restoredSlice := reflect.New(valueType).Elem()
		if sliceLen == 0 {
			return restoredSlice.Interface(), nil
		}
		for i := 0; i < int(sliceLen); i++ {
			elem, err := m.deserialize(valueType.Elem(), minSliceLen, maxSliceLen, lenPrefixType, buffer)
			if err != nil {
				return nil, err
			}
			restoredSlice = reflect.Append(restoredSlice, reflect.ValueOf(elem))
		}
		return restoredSlice.Interface(), nil
	case reflect.Map:
		mapSize, err := readLen(lenPrefixType, buffer)
		if err != nil {
			return nil, err
		}
		restoredMap := reflect.MakeMap(valueType)
		if mapSize == 0 {
			return restoredMap.Interface(), nil
		}
		for i := 0; i < int(mapSize); i++ {
			k, err := m.deserialize(valueType.Key(), minSliceLen, maxSliceLen, lenPrefixType, buffer)
			if err != nil {
				return nil, err
			}
			v, err := m.deserialize(valueType.Elem(), minSliceLen, maxSliceLen, lenPrefixType, buffer)
			if err != nil {
				return nil, err
			}
			restoredMap.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
		}
		return restoredMap.Interface(), nil
	case reflect.Ptr:
		nilPointer, err := buffer.ReadByte()
		if err != nil {
			return nil, err
		}
		// if pointer prefix is 0 then set value of the pointer to nil
		if nilPointer == 0 {
			return nil, nil
		} else {
			p := reflect.New(valueType.Elem())
			de, err := m.deserialize(valueType.Elem(), minSliceLen, maxSliceLen, lenPrefixType, buffer)
			if err != nil {
				return nil, err
			}
			p.Elem().Set(reflect.ValueOf(de))
			return p.Interface(), nil
		}
	case reflect.Struct:
		s, err := m.deserializeStruct(valueType, minSliceLen, maxSliceLen, lenPrefixType, buffer)
		if err != nil {
			return nil, err
		}
		return s, nil

	case reflect.Interface:
		nilPointer, err := buffer.ReadByte()
		if err != nil {
			return nil, err
		}
		// if pointer prefix is 0 then set value of the pointer to nil
		if nilPointer == 0 {
			return nil, nil
		} else {
			encodedType, err := buffer.ReadUint32()
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
			return m.deserialize(implementationType, minSliceLen, maxSliceLen, lenPrefixType, buffer)
		}
	}

	return nil, nil
}

func (m *SerializationManager) deserializeStruct(structType reflect.Type, minSliceLen, maxSliceLen int, lenPrefixType types.BasicKind, buffer *marshalutil.MarshalUtil) (interface{}, error) {

	if structType == reflect.TypeOf(time.Time{}) {
		restoredTime, err := buffer.ReadTime()
		if err != nil {
			return nil, err
		}
		return restoredTime, nil
	} else if structType == reflect.TypeOf(orderedmap.OrderedMap{}) {
		restoredMap := orderedmap.New()
		orderedMapSize, err := readLen(lenPrefixType, buffer)
		if err != nil {
			return nil, err
		}
		if orderedMapSize == 0 {
			return *restoredMap, nil
		}
		encodedKeyType, err := buffer.ReadUint32()
		if err != nil {
			return nil, err
		}
		encodedValueType, err := buffer.ReadUint32()
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

		for i := 0; i < orderedMapSize; i++ {
			key, err := m.deserialize(keyType, minSliceLen, maxSliceLen, lenPrefixType, buffer)
			if err != nil {
				return nil, err
			}
			value, err := m.deserialize(valueType, minSliceLen, maxSliceLen, lenPrefixType, buffer)
			if err != nil {
				return nil, err
			}
			restoredMap.Set(key, value)
		}
		return *restoredMap, nil
	} else {
		restoredStruct := reflect.New(structType).Elem()
		serializedFields, err := m.fieldCache.GetFields(structType)
		if err != nil {
			return nil, err
		}
		for _, fieldMeta := range serializedFields {
			field := structType.Field(fieldMeta.idx)
			if !fieldMeta.unpack {
				fieldValue, err := m.deserialize(field.Type, fieldMeta.minSliceLength, fieldMeta.minSliceLength, fieldMeta.lengthPrefixType, buffer)
				if err != nil {
					return nil, err
				}
				if fieldValue != nil {
					restoredStruct.Field(fieldMeta.idx).Set(reflect.ValueOf(fieldValue).Convert(field.Type))
				}
			} else {
				if !field.Anonymous {
					panic(fmt.Sprintf("unpack on non anonymous field %s", field.Name))
				}

				if field.Type.Kind() != reflect.Struct {
					panic(fmt.Sprintf("unpack on non anonymous struct field %s", field.Name))
				}

				anonEmbeddedStructType := field.Type
				anonEmbeddedSerializedFields, err := m.fieldCache.GetFields(anonEmbeddedStructType)
				if err != nil {
					return nil, err
				}
				for _, embFieldMeta := range anonEmbeddedSerializedFields {
					embStructField := anonEmbeddedStructType.Field(embFieldMeta.idx)
					embStructFieldVal, err := m.deserialize(embStructField.Type, fieldMeta.minSliceLength, fieldMeta.minSliceLength, fieldMeta.lengthPrefixType, buffer)
					if err != nil {
						return nil, err
					}
					if embStructFieldVal != nil {
						restoredStruct.Field(fieldMeta.idx).Field(embFieldMeta.idx).Set(reflect.ValueOf(embStructFieldVal).Convert(embStructField.Type))
					}
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
		if err := m.serialize(reflect.ValueOf(s).Elem(), defaultMinLen, defaultMaxLen, defaultLenPrefixType, buffer); err != nil {
			return nil, err
		}
		return buffer.Bytes(), nil
	}
	if err := m.serialize(reflect.ValueOf(s), defaultMinLen, defaultMaxLen, defaultLenPrefixType, buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (m *SerializationManager) serialize(value reflect.Value, minSliceLen, maxSliceLen int, lenPrefixType types.BasicKind, buffer *marshalutil.MarshalUtil) error {
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
			err = m.serialize(value.Index(i), minSliceLen, maxSliceLen, lenPrefixType, buffer)
			if err != nil {
				break
			}
		}
	case reflect.Slice:
		err = writeLen(value.Len(), lenPrefixType, buffer)
		if err != nil {
			break
		}
		for i := 0; i < value.Len(); i++ {
			err = m.serialize(value.Index(i), minSliceLen, maxSliceLen, lenPrefixType, buffer)
			if err != nil {
				break
			}
		}
	case reflect.Map:
		err = writeLen(value.Len(), lenPrefixType, buffer)
		if err != nil {
			break
		}

		keys := value.MapKeys()
		sort.Slice(keys, keyComparator(keys))
		for _, k := range keys {
			err = m.serialize(k, minSliceLen, maxSliceLen, lenPrefixType, buffer)
			if err != nil {
				break
			}
			err = m.serialize(value.MapIndex(k), minSliceLen, maxSliceLen, lenPrefixType, buffer)
			if err != nil {
				break
			}
		}
	case reflect.Ptr:
		if value.IsNil() {
			buffer.WriteByte(byte(0))
		} else {
			buffer.WriteByte(byte(1))
			err = m.serialize(value.Elem(), minSliceLen, maxSliceLen, lenPrefixType, buffer)
		}
	case reflect.Struct:
		structType := value.Type()
		switch structType {
		case reflect.TypeOf(time.Time{}):
			buffer.WriteTime(value.Interface().(time.Time))
		case reflect.TypeOf(orderedmap.OrderedMap{}):
			err = m.serializeOrderedMap(value, minSliceLen, maxSliceLen, lenPrefixType, buffer)
		default:
			err = m.serializeStruct(value, minSliceLen, maxSliceLen, lenPrefixType, buffer)
		}
	case reflect.Interface:
		if value.IsNil() {
			buffer.WriteByte(byte(0))
		} else {
			buffer.WriteByte(byte(1))
			interfaceType := reflect.TypeOf(value.Interface())
			var encodedType uint32
			encodedType, err = m.EncodeType(interfaceType)
			if err != nil {
				break
			}
			buffer.WriteUint32(encodedType)
			err = m.serialize(value.Elem(), minSliceLen, maxSliceLen, lenPrefixType, buffer)
		}
	}
	return err
}

func (m *SerializationManager) serializeStruct(value reflect.Value, minSliceLen, maxSliceLen int, lenPrefixType types.BasicKind, buffer *marshalutil.MarshalUtil) error {
	structType := value.Type()
	var err error
	serializedFields, err := m.fieldCache.GetFields(structType)
	if err != nil {
		return err
	}
	for _, fieldMeta := range serializedFields {
		fieldValue := value.Field(fieldMeta.idx)
		err = m.serialize(fieldValue, fieldMeta.minSliceLength, fieldMeta.maxSliceLength, fieldMeta.lengthPrefixType, buffer)
		if err != nil {
			break
		}
	}
	return err
}

func (m *SerializationManager) serializeOrderedMap(value reflect.Value, minSliceLen, maxSliceLen int, lenPrefixType types.BasicKind, buffer *marshalutil.MarshalUtil) error {
	orderedMap := value.Interface().(orderedmap.OrderedMap)

	err := writeLen(orderedMap.Size(), lenPrefixType, buffer)
	if err != nil {
		return err
	}

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
		err = m.serialize(reflect.ValueOf(key), minSliceLen, maxSliceLen, lenPrefixType, buffer)
		if err != nil {
			return false
		}
		err = m.serialize(reflect.ValueOf(value), minSliceLen, maxSliceLen, lenPrefixType, buffer)
		return err == nil
	})
	return err
}
func readLen(lenPrefixType types.BasicKind, buffer *marshalutil.MarshalUtil) (int, error) {
	switch lenPrefixType {
	case types.Uint8:
		lengthUint8, err := buffer.ReadUint8()
		return int(lengthUint8), err
	case types.Uint16:
		lengthUint16, err := buffer.ReadUint16()
		return int(lengthUint16), err
	case types.Uint32:
		lengthUint32, err := buffer.ReadUint32()
		return int(lengthUint32), err
	default:
		return 0, fmt.Errorf("unknown length prefix type %d", lenPrefixType)
	}
}
func writeLen(length int, lenPrefixType types.BasicKind, buffer *marshalutil.MarshalUtil) error {
	switch lenPrefixType {
	case types.Uint8:
		buffer.WriteUint8(uint8(length))
	case types.Uint16:
		buffer.WriteUint16(uint16(length))
	case types.Uint32:
		buffer.WriteUint32(uint32(length))
	default:
		return fmt.Errorf("unknown length prefix type %d", lenPrefixType)
	}
	return nil
}

func keyComparator(keys []reflect.Value) func(int, int) bool {
	// TODO: change to byte lexicographical order
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

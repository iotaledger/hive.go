package autoserializer

import (
	"encoding"
	"errors"
	"fmt"
	"go/types"
	"math"
	"reflect"
	"time"

	"github.com/iotaledger/hive.go/marshalutil"
)

var (
	binaryMarshallerType                   = reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem()
	binaryUnarshallerType                  = reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
	autoserializerBinarySerializerType     = reflect.TypeOf((*BinarySerializer)(nil)).Elem()
	autoserializerBinararyDeserializerType = reflect.TypeOf((*BinaryDeserializer)(nil)).Elem()
	defaultFieldMetadata                   = FieldMetadata{
		LengthPrefixType: types.Uint8,
		MaxSliceLength:   0,
		MinSliceLength:   0,
		AllowNil:         false,
	}
)

type BinaryDeserializer interface {
	DeserializeBytes(buffer *marshalutil.MarshalUtil, m *SerializationManager, metadata FieldMetadata) (err error)
}

type BinarySerializer interface {
	SerializeBytes(m *SerializationManager, metadata FieldMetadata) (data []byte, err error)
}

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
func (m *SerializationManager) Deserialize(s interface{}, data []byte) (err error) {
	buffer := marshalutil.New(data)
	value := reflect.ValueOf(s)
	valueType := reflect.TypeOf(s)
	if value.Kind() != reflect.Ptr {
		return errors.New("pointer type is required to perform deserialization")
	}
	var result interface{}
	// if the pointer type implements built-in deserializer, then use it. otherwise, handle it as a regular pointer type.
	if valueType != reflect.TypeOf(&time.Time{}) && (valueType.Implements(autoserializerBinararyDeserializerType) ||
		valueType.Implements(binaryUnarshallerType)) {
		result, err = m.DeserializeType(valueType, defaultFieldMetadata, buffer)
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
		value.Elem().Set(reflect.ValueOf(result).Elem())
	} else {
		result, err = m.DeserializeType(valueType.Elem(), defaultFieldMetadata, buffer)
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
	}
	return nil
}

func (m *SerializationManager) DeserializeType(valueType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (interface{}, error) {
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
		arrayLen := valueType.Len()
		if valueType.Elem().Kind() == reflect.Uint8 {
			bytes, err := buffer.ReadBytes(arrayLen)
			if err != nil {
				return nil, err
			}
			for i := 0; i < arrayLen; i++ {
				restoredArray.Index(i).Set(reflect.ValueOf(bytes[i]))
			}
		} else {
			for i := 0; i < arrayLen; i++ {
				elem, err := m.DeserializeType(valueType.Elem(), fieldMetadata, buffer)
				if err != nil {
					return nil, err
				}
				restoredArray.Index(i).Set(reflect.ValueOf(elem))
			}
		}
		return restoredArray.Interface(), nil
	case reflect.Slice:
		sliceLen, err := ReadLen(fieldMetadata.LengthPrefixType, buffer)
		if err != nil {
			return nil, err
		}
		err = ValidateLength(sliceLen, fieldMetadata.MinSliceLength, fieldMetadata.MaxSliceLength)
		if err != nil {
			return nil, err
		}
		restoredSlice := reflect.New(valueType).Elem()
		if sliceLen == 0 {
			return restoredSlice.Interface(), nil
		}

		if valueType.Elem().Kind() == reflect.Uint8 {
			restored, err := buffer.ReadBytes(sliceLen)
			if err != nil {
				return nil, err
			}
			restoredSlice = reflect.ValueOf(restored)
		} else {
			for i := 0; i < sliceLen; i++ {
				elem, err := m.DeserializeType(valueType.Elem(), fieldMetadata, buffer)
				if err != nil {
					return nil, err
				}
				restoredSlice = reflect.Append(restoredSlice, reflect.ValueOf(elem))
			}
		}
		return restoredSlice.Interface(), nil
	case reflect.Map:
		return nil, fmt.Errorf("native map type is not supported. use orderedmap instead")
	case reflect.Ptr:
		if fieldMetadata.AllowNil {
			nilPointer, err := buffer.ReadByte()
			if err != nil {
				return nil, err
			}
			// if pointer prefix is 0 then set value of the pointer to nil
			if nilPointer == 0 {
				return nil, nil
			}
		}
		p := reflect.New(valueType.Elem())

		// try to use built-in deserializer if the pointer type provides it
		processed, err := m.tryBuiltInDeserializer(p, valueType, fieldMetadata, buffer)
		if err != nil {
			return nil, err
		}
		if processed {
			return p.Interface(), nil
		}

		// deserialize as normal pointer type
		de, err := m.DeserializeType(valueType.Elem(), fieldMetadata, buffer)
		if err != nil {
			return nil, err
		}
		p.Elem().Set(reflect.ValueOf(de))
		return p.Interface(), nil
	case reflect.Struct:
		if valueType == reflect.TypeOf(time.Time{}) {
			restoredTime, err := buffer.ReadTime()
			if err != nil {
				return nil, err
			}
			return restoredTime, nil
		}

		structValue := reflect.New(valueType).Elem()

		// try to use built-in deserializer if the struct type provides it
		processed, err := m.tryBuiltInDeserializer(structValue, valueType, fieldMetadata, buffer)
		if err != nil {
			return nil, err
		}
		if processed {
			return structValue.Interface(), nil
		}

		// deserialize as normal structure
		s, err := m.deserializeStruct(structValue, valueType, buffer)
		if err != nil {
			return nil, err
		}
		return s, nil
	case reflect.Interface:
		if fieldMetadata.AllowNil {
			nilPointer, err := buffer.ReadByte()
			if err != nil {
				return nil, err
			}
			// if interface prefix is 0 then set value of the interface to nil
			if nilPointer == 0 {
				return nil, nil
			}
		}
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
		return m.DeserializeType(implementationType, fieldMetadata, buffer)
	}

	return nil, nil
}

func (m *SerializationManager) deserializeStruct(restoredStruct reflect.Value, structType reflect.Type, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	serializedFields, err := m.fieldCache.GetFields(structType)
	if err != nil {
		return nil, err
	}
	for _, fieldMeta := range serializedFields {
		field := structType.Field(fieldMeta.idx)
		if !fieldMeta.unpack {
			fieldValue, err := m.DeserializeType(field.Type, fieldMeta, buffer)
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
				embStructFieldVal, err := m.DeserializeType(embStructField.Type, fieldMeta, buffer)
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

// Serialize turns object into bytes.
func (m *SerializationManager) Serialize(s interface{}) ([]byte, error) {
	buffer := marshalutil.New()
	objectType := reflect.TypeOf(s)
	// if objectType is a pointer and the pointer type does not implement built-in serialization, unpack the pointer before serialization
	if objectType.Kind() == reflect.Ptr &&
		!objectType.Implements(autoserializerBinarySerializerType) &&
		!objectType.Implements(binaryMarshallerType) {
		if err := m.SerializeValue(reflect.ValueOf(s).Elem(), defaultFieldMetadata, buffer); err != nil {
			return nil, err
		}
		return buffer.Bytes(), nil
	}

	if err := m.SerializeValue(reflect.ValueOf(s), defaultFieldMetadata, buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (m *SerializationManager) SerializeValue(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) error {
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
		arrayLen := value.Len()
		if value.Type().Elem().Kind() == reflect.Uint8 {
			for i := 0; i < arrayLen; i++ {
				buffer.WriteByte(uint8(value.Index(i).Uint()))
			}
		} else {
			for i := 0; i < arrayLen; i++ {
				err = m.SerializeValue(value.Index(i), fieldMetadata, buffer)
				if err != nil {
					break
				}
			}
		}
	case reflect.Slice:
		err = WriteLen(value.Len(), fieldMetadata.LengthPrefixType, buffer)
		if err != nil {
			break
		}

		err = ValidateLength(value.Len(), fieldMetadata.MinSliceLength, fieldMetadata.MaxSliceLength)
		if err != nil {
			return err
		}

		if value.Type().Elem().Kind() == reflect.Uint8 {
			buffer.WriteBytes(value.Bytes())
		} else {
			for i := 0; i < value.Len(); i++ {
				err = m.SerializeValue(value.Index(i), fieldMetadata, buffer)
				if err != nil {
					break
				}
			}
		}
	case reflect.Map:
		err = fmt.Errorf("native map type is not supported. use orderedmap instead")
		break
	case reflect.Ptr:
		// write first byte only if AllowNil set to true
		if value.IsNil() && fieldMetadata.AllowNil {
			buffer.WriteByte(byte(0))
			break
		} else if fieldMetadata.AllowNil {
			buffer.WriteByte(byte(1))
		}

		valueType := reflect.TypeOf(value.Interface())

		var processed bool
		processed, err = m.tryBuiltInSerializer(value, valueType, fieldMetadata, buffer)
		if err != nil || processed {
			break
		}

		err = m.SerializeValue(value.Elem(), fieldMetadata, buffer)
	case reflect.Struct:
		valueType := value.Type()
		if valueType == reflect.TypeOf(time.Time{}) {
			buffer.WriteTime(value.Interface().(time.Time))
			break
		}
		// serialize using built-in serializer if it's available
		var processed bool
		processed, err = m.tryBuiltInSerializer(value, valueType, fieldMetadata, buffer)
		if err != nil || processed {
			break
		}

		// serialize as regular struct
		err = m.serializeStruct(value, buffer)
	case reflect.Interface:
		// write first byte only if AllowNil set to true
		if value.IsNil() && fieldMetadata.AllowNil {
			buffer.WriteByte(byte(0))
			break
		} else if fieldMetadata.AllowNil {
			buffer.WriteByte(byte(1))
		}
		interfaceType := reflect.TypeOf(value.Interface())
		var encodedType uint32
		encodedType, err = m.EncodeType(interfaceType)
		if err != nil {
			break
		}
		buffer.WriteUint32(encodedType)
		err = m.SerializeValue(value.Elem(), fieldMetadata, buffer)
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
	for _, fieldMeta := range serializedFields {
		fieldValue := value.Field(fieldMeta.idx)
		err = m.SerializeValue(fieldValue, fieldMeta, buffer)
		if err != nil {
			break
		}
	}
	return err
}

func (m *SerializationManager) tryBuiltInSerializer(value reflect.Value, valueType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (processed bool, err error) {
	if valueType.Implements(binaryMarshallerType) {
		marshaller := value.Interface().(encoding.BinaryMarshaler)
		var bytes []byte
		bytes, err = marshaller.MarshalBinary()
		if err != nil {
			return true, err
		}
		err = WriteLen(len(bytes), fieldMetadata.LengthPrefixType, buffer)
		if err != nil {
			return true, err
		}
		buffer.WriteBytes(bytes)
		return true, nil
	} else if valueType.Implements(autoserializerBinarySerializerType) {
		marshaller := value.Interface().(BinarySerializer)
		var bytes []byte
		bytes, err = marshaller.SerializeBytes(m, fieldMetadata)
		if err != nil {
			return true, err
		}

		buffer.WriteBytes(bytes)
		return true, nil
	}
	return false, err
}

func (m *SerializationManager) tryBuiltInDeserializer(p reflect.Value, structType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (processed bool, err error) {
	if structType.Implements(binaryUnarshallerType) {
		var structSize int
		structSize, err = ReadLen(fieldMetadata.LengthPrefixType, buffer)
		if err != nil {
			return true, err
		}
		structBytes, err := buffer.ReadBytes(structSize)
		if err != nil {
			return true, err
		}
		restoredStruct := p.Interface().(encoding.BinaryUnmarshaler)
		err = restoredStruct.UnmarshalBinary(structBytes)
		if err != nil {
			return true, err
		}
		return true, nil
	} else if structType.Implements(autoserializerBinararyDeserializerType) {
		restoredStruct := p.Interface().(BinaryDeserializer)
		err = restoredStruct.DeserializeBytes(buffer, m, fieldMetadata)
		if err != nil {
			return true, err
		}
		return true, nil
	}
	return
}

func ReadLen(lenPrefixType types.BasicKind, buffer *marshalutil.MarshalUtil) (int, error) {
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

func WriteLen(length int, lenPrefixType types.BasicKind, buffer *marshalutil.MarshalUtil) error {
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

func ValidateLength(length int, minSliceLen int, maxSliceLen int) (err error) {
	if length < minSliceLen {
		err = fmt.Errorf("collection is required to have at least %d elements instead of %d", minSliceLen, length)
	}
	if maxSliceLen > 0 && length > maxSliceLen {
		err = fmt.Errorf("collection is required to have at most %d elements instead of %d", maxSliceLen, length)
	}
	return
}

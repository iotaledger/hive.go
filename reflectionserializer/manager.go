package reflectionserializer

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"go/types"
	"math"
	"reflect"
	"sort"
	"time"

	"github.com/iotaledger/hive.go/marshalutil"
	customtypes "github.com/iotaledger/hive.go/types"
	"github.com/iotaledger/hive.go/typeutils"
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

// BinaryDeserializer interface is used to implement built-in deserialization of complex structures, usually collections.
type BinaryDeserializer interface {
	DeserializeBytes(buffer *marshalutil.MarshalUtil, m *SerializationManager, metadata FieldMetadata) (err error)
}

// BinarySerializer interface is used to implement built-in serialization of complex structures, usually collections.
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
		fieldCache:   newFieldCache(),
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
	var done bool
	// if the pointer type implements built-in deserializer, then use it. otherwise, handle it as a regular pointer type.
	if valueType != reflect.TypeOf(&time.Time{}) && (valueType.Implements(autoserializerBinararyDeserializerType) ||
		valueType.Implements(binaryUnarshallerType)) {
		result, err = m.DeserializeType(valueType, defaultFieldMetadata, buffer)
		if err != nil {
			return err
		}
		done, err = buffer.DoneReading()
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
		done, err = buffer.DoneReading()
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

// DeserializeType deserializes give type from the buffer according to FieldMetadata.
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
		tmp, err := ReadLen(fieldMetadata.LengthPrefixType, buffer)
		if err != nil {
			return nil, err
		}
		bytesRead, err := buffer.ReadBytes(tmp)
		if err != nil {
			return nil, err
		}
		restoredString := string(bytesRead)
		return restoredString, nil
	case reflect.Array:
		return m.deserializeArray(valueType, fieldMetadata, buffer)
	case reflect.Slice:
		return m.deserializeSlice(valueType, fieldMetadata, buffer)
	case reflect.Map:
		return nil, fmt.Errorf("native map type is not supported. use orderedmap instead")
	case reflect.Ptr:
		return m.deserializePointer(valueType, fieldMetadata, buffer)
	case reflect.Struct:
		return m.deserializeStructure(valueType, fieldMetadata, buffer)
	case reflect.Interface:
		return m.deserializeInterface(valueType, fieldMetadata, buffer)
	}
	return nil, nil
}

func (m *SerializationManager) deserializeInterface(valueType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	// if interface value can be nil, write first byte to indicate whether it has value
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

func (m *SerializationManager) deserializeStructure(valueType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	// handle time struct individually
	if valueType == reflect.TypeOf(time.Time{}) {
		restoredTime, err := buffer.ReadTime()
		if err != nil {
			return nil, err
		}
		return restoredTime, nil
	}

	structValue := reflect.New(valueType).Elem()

	// if the struct type provides it try to use built-in deserializer
	processed, err := m.tryBuiltInDeserializer(structValue, valueType, fieldMetadata, buffer)
	if err != nil {
		return nil, err
	}
	if processed {
		return structValue.Interface(), nil
	}

	// deserialize as normal structure
	s, err := m.deserializeFields(structValue, valueType, buffer)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (m *SerializationManager) deserializeFields(restoredStruct reflect.Value, structType reflect.Type, buffer *marshalutil.MarshalUtil) (interface{}, error) {
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

func (m *SerializationManager) deserializePointer(valueType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	// if pointer value can be nil, write first byte to indicate whether it has value
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
}

func (m *SerializationManager) deserializeSlice(valueType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	// read length of the slice and validate that it matches the bounds
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
		// special handling of byte slice to optimize execution
		restored, err := buffer.ReadBytes(sliceLen)
		if err != nil {
			return nil, err
		}
		restoredSlice = reflect.ValueOf(restored)
	} else if fieldMetadata.LexicalOrder || fieldMetadata.NoDuplicates {
		// if lexical ordering or no duplicates is required, perform additional processing
		readOffset := buffer.ReadOffset()
		elements := make([]struct {
			elem      interface{}
			elemBytes []byte
		}, 0)
		elementsMap := make(map[string]customtypes.Empty)
		for i := 0; i < sliceLen; i++ {
			elem, err := m.DeserializeType(valueType.Elem(), fieldMetadata, buffer)
			if err != nil {
				return nil, err
			}
			bytesRead := buffer.ReadOffset() - readOffset
			buffer.ReadSeek(-bytesRead)
			elemBytes, err := buffer.ReadBytes(bytesRead)
			readOffset = buffer.ReadOffset()

			elemBytesString := typeutils.BytesToString(elemBytes)
			if err != nil {
				return nil, err
			}

			if _, seenAlready := elementsMap[elemBytesString]; fieldMetadata.NoDuplicates && seenAlready {
				continue
			}

			elements = append(elements, struct {
				elem      interface{}
				elemBytes []byte
			}{elem, elemBytes})
			elementsMap[elemBytesString] = customtypes.Void
		}
		if fieldMetadata.LexicalOrder {
			// sort inputs
			sort.Slice(elements, func(i, j int) bool {
				return bytes.Compare(elements[i].elemBytes, elements[j].elemBytes) < 0
			})
		}
		for _, sortedInput := range elements {
			restoredSlice = reflect.Append(restoredSlice, reflect.ValueOf(sortedInput.elem))
		}

	} else {
		// simply deserialize the slice
		for i := 0; i < sliceLen; i++ {
			elem, err := m.DeserializeType(valueType.Elem(), fieldMetadata, buffer)
			if err != nil {
				return nil, err
			}
			restoredSlice = reflect.Append(restoredSlice, reflect.ValueOf(elem))
		}
	}
	return restoredSlice.Interface(), nil
}

func (m *SerializationManager) deserializeArray(valueType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	restoredArray := reflect.New(valueType).Elem()
	arrayLen := valueType.Len()
	if valueType.Elem().Kind() == reflect.Uint8 {
		// special handling of byte array to optimize execution
		bytesRead, err := buffer.ReadBytes(arrayLen)
		if err != nil {
			return nil, err
		}
		for i := 0; i < arrayLen; i++ {
			restoredArray.Index(i).Set(reflect.ValueOf(bytesRead[i]))
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
}

func (m *SerializationManager) tryBuiltInDeserializer(p reflect.Value, structType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (processed bool, err error) {
	if structType.Implements(binaryUnarshallerType) {
		// if the struct implements encoding.BinaryUnmarshaller
		// read number of bytes the struct was serialized to
		var structSize int
		var structBytes []byte
		structSize, err = ReadLen(fieldMetadata.LengthPrefixType, buffer)
		if err != nil {
			return true, err
		}
		structBytes, err = buffer.ReadBytes(structSize)
		if err != nil {
			return true, err
		}
		// unmarshal the structure using read bytes
		restoredStruct := p.Interface().(encoding.BinaryUnmarshaler)
		err = restoredStruct.UnmarshalBinary(structBytes)
		if err != nil {
			return true, err
		}
		return true, nil
	} else if structType.Implements(autoserializerBinararyDeserializerType) {
		// if the structure implements BinaryDeserializer
		// number of bytes is not needed, because if the structure has been serialized correctly,
		// it will deserialize correctly and use only as many bytes from the buffer as needed, leaving others untouched
		restoredStruct := p.Interface().(BinaryDeserializer)
		err = restoredStruct.DeserializeBytes(buffer, m, fieldMetadata)
		if err != nil {
			return true, err
		}
		return true, nil
	}
	return
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

// SerializeValue serializes given value into the buffer according to FieldMetadata
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
		err = WriteLen(len(value.String()), fieldMetadata.LengthPrefixType, buffer)
		if err == nil {
			buffer.WriteBytes([]byte(value.String()))
		}
	case reflect.Array:
		err = m.serializeArray(value, fieldMetadata, buffer)
	case reflect.Slice:
		err = m.serializeSlice(value, fieldMetadata, buffer)
	case reflect.Map:
		err = fmt.Errorf("native map type is not supported. use orderedmap instead")
	case reflect.Ptr:
		err = m.serializePointer(value, fieldMetadata, buffer)
	case reflect.Struct:
		err = m.serializeStructure(value, fieldMetadata, buffer)
	case reflect.Interface:
		err = m.serializeInterface(value, fieldMetadata, buffer)
	}
	return err
}

func (m *SerializationManager) serializeInterface(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) error {
	// write first byte only if AllowNil set to true
	if value.IsNil() && fieldMetadata.AllowNil {
		buffer.WriteByte(byte(0))
		return nil
	} else if fieldMetadata.AllowNil {
		buffer.WriteByte(byte(1))
	}

	interfaceType := reflect.TypeOf(value.Interface())
	encodedType, err := m.EncodeType(interfaceType)
	if err != nil {
		return err
	}
	buffer.WriteUint32(encodedType)
	err = m.SerializeValue(value.Elem(), fieldMetadata, buffer)
	return err
}

func (m *SerializationManager) serializeStructure(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) error {
	valueType := value.Type()
	// individual serialization for time.Time
	if valueType == reflect.TypeOf(time.Time{}) {
		buffer.WriteTime(value.Interface().(time.Time))
		return nil
	} else if processed, err := m.tryBuiltInSerializer(value, valueType, fieldMetadata, buffer); err != nil || processed {
		// serialize using built-in serializer if it's available
		return err
	} else {
		// serialize as regular struct
		err = m.serializeFields(value, buffer)
		return err
	}
}

func (m *SerializationManager) serializeFields(value reflect.Value, buffer *marshalutil.MarshalUtil) error {
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

func (m *SerializationManager) serializePointer(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) error {
	// write first byte only if AllowNil set to true
	if value.IsNil() && fieldMetadata.AllowNil {
		buffer.WriteByte(byte(0))
		return nil
	} else if fieldMetadata.AllowNil {
		buffer.WriteByte(byte(1))
	}

	// if pointer implements built-in serialization, use it
	valueType := reflect.TypeOf(value.Interface())
	processed, err := m.tryBuiltInSerializer(value, valueType, fieldMetadata, buffer)
	if err != nil || processed {
		return err
	}

	// or else serialize as regular pointer
	err = m.SerializeValue(value.Elem(), fieldMetadata, buffer)
	return err
}

func (m *SerializationManager) serializeSlice(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) error {
	if value.Type().Elem().Kind() == reflect.Uint8 {
		// serialize slice of bytes individually to optimize execution
		err := WriteLen(value.Len(), fieldMetadata.LengthPrefixType, buffer)
		if err != nil {
			return err
		}

		err = ValidateLength(value.Len(), fieldMetadata.MinSliceLength, fieldMetadata.MaxSliceLength)
		if err != nil {
			return err
		}
		buffer.WriteBytes(value.Bytes())
	} else if fieldMetadata.LexicalOrder || fieldMetadata.NoDuplicates {
		// if lexical order or no duplicates is required, perform necessary preprocessing to sort or remove duplicates
		elems := make([][]byte, 0)
		elemsMap := make(map[string]customtypes.Empty)

		for i := 0; i < value.Len(); i++ {
			elemBuffer := marshalutil.New()
			err := m.SerializeValue(value.Index(i), fieldMetadata, elemBuffer)
			if err != nil {
				return err
			}
			elemBytes := elemBuffer.Bytes()
			elemBytesString := typeutils.BytesToString(elemBytes)
			if _, seenElem := elemsMap[elemBytesString]; fieldMetadata.NoDuplicates && seenElem {
				continue
			}
			elems = append(elems, elemBytes)
			elemsMap[elemBytesString] = customtypes.Void
		}

		if fieldMetadata.LexicalOrder {
			// sort inputs
			sort.Slice(elems, func(i, j int) bool {
				return bytes.Compare(elems[i], elems[j]) < 0
			})
		}

		// validate length of the slice after preprocessing
		err := ValidateLength(len(elems), fieldMetadata.MinSliceLength, fieldMetadata.MaxSliceLength)
		if err != nil {
			return err
		}

		// write new length of the slice and its elems to the buffer
		err = WriteLen(len(elems), fieldMetadata.LengthPrefixType, buffer)
		if err != nil {
			return err
		}

		for _, sortedElem := range elems {
			buffer.WriteBytes(sortedElem)
		}
	} else {
		// validate slice length according to specified values
		err := ValidateLength(value.Len(), fieldMetadata.MinSliceLength, fieldMetadata.MaxSliceLength)
		if err != nil {
			return err
		}

		// write slice length and its elems
		err = WriteLen(value.Len(), fieldMetadata.LengthPrefixType, buffer)
		if err != nil {
			return err
		}

		for i := 0; i < value.Len(); i++ {
			err = m.SerializeValue(value.Index(i), fieldMetadata, buffer)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *SerializationManager) serializeArray(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) error {
	arrayLen := value.Len()
	if value.Type().Elem().Kind() == reflect.Uint8 {
		// individual serialization for byte arrays in order to optimize execution
		for i := 0; i < arrayLen; i++ {
			buffer.WriteByte(uint8(value.Index(i).Uint()))
		}
	} else {
		for i := 0; i < arrayLen; i++ {
			err := m.SerializeValue(value.Index(i), fieldMetadata, buffer)
			if err != nil {
				break
			}
		}
	}
	return nil
}

func (m *SerializationManager) tryBuiltInSerializer(value reflect.Value, valueType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (processed bool, err error) {
	if valueType.Implements(binaryMarshallerType) {
		// if the struct implements encoding.BinaryUnmarshaller
		// serialize using built-in method
		marshaller := value.Interface().(encoding.BinaryMarshaler)
		var bytesMarshalled []byte
		bytesMarshalled, err = marshaller.MarshalBinary()
		if err != nil {
			return true, err
		}

		// write number of bytes the struct was serialized to
		err = WriteLen(len(bytesMarshalled), fieldMetadata.LengthPrefixType, buffer)
		if err != nil {
			return true, err
		}
		// write byte slice of serialized structure
		buffer.WriteBytes(bytesMarshalled)
		return true, nil
	} else if valueType.Implements(autoserializerBinarySerializerType) {
		// if the structure implements BinarySerializer
		// writing number of bytes, because if the structure has been serialized correctly,
		// it will deserialize correctly and use only as many bytes from the buffer as needed, leaving others untouched
		marshaller := value.Interface().(BinarySerializer)
		var bytesMarshalled []byte
		bytesMarshalled, err = marshaller.SerializeBytes(m, fieldMetadata)
		if err != nil {
			return true, err
		}

		buffer.WriteBytes(bytesMarshalled)
		return true, nil
	}
	return false, err
}

// ReadLen reads length of a collection from the buffer according to lenPrefixType
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

// WriteLen writes length of a collection from the buffer according to lenPrefixType
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

// ValidateLength is used to make sure that the length of a collection is within bounds specified in struct tags.
func ValidateLength(length int, minSliceLen int, maxSliceLen int) (err error) {
	if length < minSliceLen {
		err = fmt.Errorf("collection is required to have at least %d elements instead of %d", minSliceLen, length)
	}
	if maxSliceLen > 0 && length > maxSliceLen {
		err = fmt.Errorf("collection is required to have at most %d elements instead of %d", maxSliceLen, length)
	}
	return
}

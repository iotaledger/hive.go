package refseri

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"go/types"
	"math"
	"reflect"
	"time"

	"github.com/iotaledger/hive.go/marshalutil"
	customtypes "github.com/iotaledger/hive.go/types"
	"github.com/iotaledger/hive.go/typeutils"
)

var (
	binaryMarshallerType               = reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem()
	binaryUnarshallerType              = reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
	autoserializerBinarySerializerType = reflect.TypeOf((*BinarySerializer)(nil)).Elem()
	binaryDeserializerType             = reflect.TypeOf((*BinaryDeserializer)(nil)).Elem()
	defaultFieldMetadata               = FieldMetadata{
		LengthPrefixType: types.Uint8,
		MaxSliceLength:   0,
		MinSliceLength:   0,
		AllowNil:         false,
	}
)

var ErrNotAllBytesRead = errors.New("did not read all bytes from the buffer")
var ErrMapNotSupported = errors.New("native map type is not supported. use orderedmap instead")
var ErrSerializeInterface = errors.New("couldn't deserialize interface")
var ErrUnpackAnonymous = errors.New("cannot unpack on non anonymous field")
var ErrUnpackNonStruct = errors.New("cannot unpack on non struct field")
var ErrUnknownLengthPrefix = errors.New("unknown length prefix type")
var ErrNilNotAllowed = errors.New("nil value is not allowed")
var ErrNaNValue = errors.New("NaN float value")
var ErrSliceMinLength = errors.New("collection is required to have min number of elements")
var ErrSliceMaxLength = errors.New("collection is required to have max number of elements")
var ErrLexicalOrderViolated = errors.New("lexical order violated")
var ErrNoDuplicatesViolated = errors.New("no duplicates requirement violated")

// BinaryDeserializer interface is used to implement built-in deserialization of complex structures, usually collections.
type BinaryDeserializer interface {
	DeserializeBytes(buffer *marshalutil.MarshalUtil, m *Serializer, metadata FieldMetadata) (err error)
}

// BinarySerializer interface is used to implement built-in serialization of complex structures, usually collections.
type BinarySerializer interface {
	SerializeBytes(m *Serializer, metadata FieldMetadata) (data []byte, err error)
}

// Serializer stores TypeRegistry and is used to automatically serialize and deserialize structures.
type Serializer struct {
	*TypeRegistry
	*fieldCache
}

// NewSerializationManager creates new serialization manager with default TypeRegistry.
func NewSerializationManager() *Serializer {
	sm := &Serializer{
		TypeRegistry: NewTypeRegistry(),
		fieldCache:   newFieldCache(),
	}

	return sm
}

// Deserialize data into type s and save its value. s must be a pointer type.
func (m *Serializer) Deserialize(s interface{}, data []byte) (err error) {
	buffer := marshalutil.New(data)
	value := reflect.ValueOf(s)
	valueType := reflect.TypeOf(s)
	if value.Kind() != reflect.Ptr {
		err = fmt.Errorf("pointer type is required to perform deserialization instead of %s", value.Kind())
		return
	}
	var result interface{}
	var done bool
	// if the pointer type implements built-in deserializer, then use it. otherwise, handle it as a regular pointer type.
	isTimeType := valueType == reflect.TypeOf(&time.Time{})
	implementsBinaryDeserializer := valueType.Implements(binaryDeserializerType)
	implementsBinaryUnmarshaller := valueType.Implements(binaryUnarshallerType)
	var deserializeType reflect.Type

	// unpack the pointer when the type does not implement built-in serialization, otherwise use the pointer type
	if !isTimeType && (implementsBinaryDeserializer || implementsBinaryUnmarshaller) {
		deserializeType = valueType
	} else {
		deserializeType = valueType.Elem()
	}
	result, err = m.DeserializeType(deserializeType, defaultFieldMetadata, buffer)
	if err != nil {
		return
	}
	done, err = buffer.DoneReading()
	if err != nil {
		return
	}
	if !done {
		return ErrNotAllBytesRead
	}

	// unpack the returned pointer value when the type implements built-in serialization, otherwise use the returned value
	var deserializedValue reflect.Value
	if !isTimeType && (implementsBinaryDeserializer || implementsBinaryUnmarshaller) {
		deserializedValue = reflect.ValueOf(result).Elem()
	} else {
		deserializedValue = reflect.ValueOf(result)
	}
	value.Elem().Set(deserializedValue)

	return
}

// DeserializeType deserializes the given type from the buffer according to FieldMetadata.
func (m *Serializer) DeserializeType(valueType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	switch valueType.Kind() {
	case reflect.Bool:
		val, err := buffer.ReadBool()
		if err != nil {
			return nil, err
		}
		return val, nil
	case reflect.Int8:
		val, err := buffer.ReadInt8()
		if err != nil {
			return nil, err
		}
		return val, nil
	case reflect.Int16:
		val, err := buffer.ReadInt16()
		if err != nil {
			return nil, err
		}
		return val, nil
	case reflect.Int32:
		val, err := buffer.ReadInt32()
		if err != nil {
			return nil, err
		}
		return val, nil
	case reflect.Int64:
		val, err := buffer.ReadInt64()
		if err != nil {
			return nil, err
		}
		return val, nil
	case reflect.Int:
		val, err := buffer.ReadInt64()
		if err != nil {
			return nil, err
		}
		return int(val), nil
	case reflect.Uint8:
		val, err := buffer.ReadUint8()
		if err != nil {
			return nil, err
		}
		return val, nil
	case reflect.Uint16:
		val, err := buffer.ReadUint16()
		if err != nil {
			return nil, err
		}
		return val, nil
	case reflect.Uint32:
		val, err := buffer.ReadUint32()
		if err != nil {
			return nil, err
		}
		return val, nil
	case reflect.Uint64:
		val, err := buffer.ReadUint64()
		if err != nil {
			return nil, err
		}
		return val, nil
	case reflect.Uint:
		val, err := buffer.ReadUint64()
		if err != nil {
			return nil, err
		}
		return uint(val), nil
	case reflect.Float32:
		val, err := buffer.ReadFloat32()
		if err != nil {
			return nil, err
		}
		return val, nil
	case reflect.Float64:
		val, err := buffer.ReadFloat64()
		if err != nil {
			return nil, err
		}
		return val, nil
	case reflect.String:
		strLen, err := ReadLen(fieldMetadata.LengthPrefixType, buffer)
		if err != nil {
			return nil, err
		}
		bytesRead, err := buffer.ReadBytes(strLen)
		if err != nil {
			return nil, err
		}
		restoredString := string(bytesRead)
		return restoredString, nil
	case reflect.Array:
		return m.deserializeArray(valueType, buffer, fieldMetadata)
	case reflect.Slice:
		return m.deserializeSlice(valueType, fieldMetadata, buffer)
	case reflect.Map:
		return nil, ErrMapNotSupported
	case reflect.Ptr:
		return m.deserializePointer(valueType, fieldMetadata, buffer)
	case reflect.Struct:
		return m.deserializeStructure(valueType, fieldMetadata, buffer)
	case reflect.Interface:
		return m.deserializeInterface(valueType, fieldMetadata, buffer)
	}
	return nil, nil
}

func (m *Serializer) deserializeInterface(valueType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	// check if pointer can be nil and check if it is
	isNil, err := checkIfNil(fieldMetadata, buffer)
	if err != nil || isNil {
		return nil, err
	}

	encodedType, err := buffer.ReadUint32()
	if err != nil {
		err = fmt.Errorf("%w: error while reading encoded type", err)
		return nil, err
	}
	implementationType, err := m.DecodeType(encodedType)
	if err != nil {
		err = fmt.Errorf("%w: error while decoding type", err)
		return nil, err
	}
	if !implementationType.Implements(valueType) {
		return nil, fmt.Errorf("%w: %s must implement interface %s", ErrSerializeInterface, implementationType, valueType)
	}
	return m.DeserializeType(implementationType, fieldMetadata, buffer)
}

func (m *Serializer) deserializeStructure(valueType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	// handle time struct individually
	if valueType == reflect.TypeOf(time.Time{}) {
		restoredTime, err := buffer.ReadTime()
		if err != nil {
			err = fmt.Errorf("%w: error deserializing time", err)
			return nil, err
		}
		return restoredTime, nil
	}

	structValue := reflect.New(valueType).Elem()

	// if the struct type provides it try to use built-in deserializer
	processed, err := m.tryBuiltInDeserializer(structValue, valueType, fieldMetadata, buffer)
	if err != nil {
		err = fmt.Errorf("%w: error while using built-in deserializer", err)
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

func (m *Serializer) deserializeFields(restoredStruct reflect.Value, structType reflect.Type, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	serializedFields, err := m.fieldCache.Fields(structType)
	if err != nil {
		err = fmt.Errorf("%w: error retrieving struct fields", err)
		return nil, err
	}
	for _, fieldMeta := range serializedFields {
		field := structType.Field(fieldMeta.Idx)

		if !fieldMeta.Unpack {
			fieldValue, err := m.DeserializeType(field.Type, fieldMeta, buffer)
			if err != nil {
				err = fmt.Errorf("%w: error deserializing field: '%s'", err, fieldMeta.Name)
				return nil, err
			}
			if fieldValue != nil {
				restoredStruct.Field(fieldMeta.Idx).Set(reflect.ValueOf(fieldValue).Convert(field.Type))
			}
			continue
		}

		if !field.Anonymous {
			err = fmt.Errorf("%w: '%s'", ErrUnpackAnonymous, fieldMeta.Name)
			return nil, err
		}

		if field.Type.Kind() != reflect.Struct {
			err = fmt.Errorf("%w: '%s'", ErrUnpackNonStruct, fieldMeta.Name)
			return nil, err
		}

		anonEmbeddedStructType := field.Type
		anonEmbeddedSerializedFields, err := m.fieldCache.Fields(anonEmbeddedStructType)
		if err != nil {
			err = fmt.Errorf("%w: error retrieving struct fields to unpack field '%s'", err, fieldMeta.Name)
			return nil, err
		}
		for _, embFieldMeta := range anonEmbeddedSerializedFields {
			embStructField := anonEmbeddedStructType.Field(embFieldMeta.Idx)
			embStructFieldVal, err := m.DeserializeType(embStructField.Type, fieldMeta, buffer)
			if err != nil {
				err = fmt.Errorf("%w: error deserializing: '%s'", err, fieldMeta.Name)
				return nil, err
			}
			if embStructFieldVal != nil {
				restoredStruct.Field(fieldMeta.Idx).Field(embFieldMeta.Idx).Set(reflect.ValueOf(embStructFieldVal).Convert(embStructField.Type))
			}
		}

	}
	return restoredStruct.Interface(), nil
}

func (m *Serializer) deserializePointer(valueType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	// check if pointer can be nil and check if it is
	isNil, err := checkIfNil(fieldMetadata, buffer)
	if err != nil || isNil {
		return nil, err
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

func (m *Serializer) deserializeSlice(valueType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	// read length of the slice and validate that it matches the bounds
	sliceLen, err := ReadLen(fieldMetadata.LengthPrefixType, buffer)
	if err != nil {
		err = fmt.Errorf("%w: error while reading slice length", err)
		return nil, err
	}

	if err = ValidateLength(sliceLen, fieldMetadata.MinSliceLength, fieldMetadata.MaxSliceLength); err != nil {
		return nil, err
	}
	if sliceLen == 0 {
		restoredSlice := reflect.New(valueType).Elem()
		return restoredSlice.Interface(), nil
	}
	if valueType.Elem().Kind() == reflect.Uint8 {
		// special handling of byte slice to optimize execution
		return m.deserializeSliceOfBytes(sliceLen, buffer)
	} else if fieldMetadata.LexicalOrder || fieldMetadata.NoDuplicates {
		// if lexical ordering or no duplicates is required, perform additional processing
		return m.deserializeConstrainedSlice(valueType, sliceLen, fieldMetadata, buffer)
	}
	// simply deserialize the slice
	return m.deserializeRegularSlice(valueType, sliceLen, fieldMetadata, buffer)

}

func (m *Serializer) deserializeRegularSlice(valueType reflect.Type, sliceLen int, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	restoredSlice := reflect.New(valueType).Elem()

	for i := 0; i < sliceLen; i++ {
		elem, err := m.DeserializeType(valueType.Elem(), fieldMetadata, buffer)
		if err != nil {
			return nil, err
		}
		restoredSlice = reflect.Append(restoredSlice, reflect.ValueOf(elem))
	}
	return restoredSlice.Interface(), nil
}

func (m *Serializer) deserializeConstrainedSlice(valueType reflect.Type, sliceLen int, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	restoredSlice := reflect.New(valueType).Elem()

	readOffset := buffer.ReadOffset()
	elementsMap := make(map[string]customtypes.Empty)
	var prevBytes []byte
	for i := 0; i < sliceLen; i++ {
		// deserialize slice element
		elem, err := m.DeserializeType(valueType.Elem(), fieldMetadata, buffer)
		if err != nil {
			err = fmt.Errorf("%w: error while deserializing slice elem on position %d", err, i)
			return reflect.Value{}, err
		}

		// read raw bytes again to check for duplicates and lexical order
		bytesRead := buffer.ReadOffset() - readOffset
		buffer.ReadSeek(-bytesRead)
		elemBytes, err := buffer.ReadBytes(bytesRead)
		readOffset = buffer.ReadOffset()

		// check if element is not duplicate of already deserialized element. return error otherwise
		if fieldMetadata.NoDuplicates {
			elemBytesString := typeutils.BytesToString(elemBytes)
			if err != nil {
				err = fmt.Errorf("%w: error while writing byte slice as string", err)
				return reflect.Value{}, err
			}

			if _, seenAlready := elementsMap[elemBytesString]; seenAlready {
				err = fmt.Errorf("%w: slice index %d", ErrNoDuplicatesViolated, i)
				return reflect.Value{}, err
			}
			elementsMap[elemBytesString] = customtypes.Void
		}

		// check if lexical order is preserved, return error otherwise
		if fieldMetadata.LexicalOrder && i > 0 {
			elemsInOrder := bytes.Compare(prevBytes, elemBytes) < 0
			if !elemsInOrder {
				err = fmt.Errorf("%w: slice index %d", ErrLexicalOrderViolated, i)
				return reflect.Value{}, err
			}
		}
		prevBytes = elemBytes

		// append element to resulting slice
		restoredSlice = reflect.Append(restoredSlice, reflect.ValueOf(elem))
	}

	return restoredSlice.Interface(), nil
}

func (m *Serializer) deserializeSliceOfBytes(sliceLen int, buffer *marshalutil.MarshalUtil) (interface{}, error) {
	restored, err := buffer.ReadBytes(sliceLen)
	if err != nil {
		err = fmt.Errorf("%w: error while reading slice of bytes", err)
		return nil, err
	}
	return restored, nil
}

func (m *Serializer) deserializeArray(valueType reflect.Type, buffer *marshalutil.MarshalUtil, fieldMetadata FieldMetadata) (interface{}, error) {
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
		return restoredArray.Interface(), nil
	}
	for i := 0; i < arrayLen; i++ {
		elem, err := m.DeserializeType(valueType.Elem(), fieldMetadata, buffer)
		if err != nil {
			return nil, err
		}
		restoredArray.Index(i).Set(reflect.ValueOf(elem))
	}
	return restoredArray.Interface(), nil
}

func (m *Serializer) tryBuiltInDeserializer(p reflect.Value, structType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (processed bool, err error) {
	if structType.Implements(binaryUnarshallerType) {
		processed = true
		err = m.deserializeBinaryUnmarshaller(p, fieldMetadata, buffer)
		return
	} else if structType.Implements(binaryDeserializerType) {
		processed = true
		err = m.deserializeBinaryDeserializer(p, fieldMetadata, buffer)
		return
	}
	return
}

func (m *Serializer) deserializeBinaryDeserializer(p reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (err error) {
	// if the structure implements BinaryDeserializer
	// number of bytes is not needed, because if the structure has been serialized correctly,
	// it will deserialize correctly and use only as many bytes from the buffer as needed, leaving others untouched
	restoredStruct := p.Interface().(BinaryDeserializer)
	if err = restoredStruct.DeserializeBytes(buffer, m, fieldMetadata); err != nil {
		err = fmt.Errorf("%w: error while deserializing struct", err)
		return
	}
	return
}

func (m *Serializer) deserializeBinaryUnmarshaller(p reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (err error) {
	// if the struct implements encoding.BinaryUnmarshaller
	// read number of bytes the struct was serialized to
	var structSize int
	var structBytes []byte
	structSize, err = ReadLen(fieldMetadata.LengthPrefixType, buffer)
	if err != nil {
		err = fmt.Errorf("%w: error while reading struct length", err)
		return
	}
	structBytes, err = buffer.ReadBytes(structSize)
	if err != nil {
		err = fmt.Errorf("%w: error while reading struct bytes", err)
		return
	}
	// unmarshal the structure using read bytes
	restoredStruct := p.Interface().(encoding.BinaryUnmarshaler)

	if err = restoredStruct.UnmarshalBinary(structBytes); err != nil {
		err = fmt.Errorf("%w: error while deserializing struct", err)
		return
	}
	return
}

// Serialize turns object into bytes.
func (m *Serializer) Serialize(s interface{}) ([]byte, error) {
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
func (m *Serializer) SerializeValue(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) error {
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
			err = ErrNaNValue
			break
		}
		buffer.WriteFloat32(float32(value.Float()))
	case reflect.Float64:
		if math.IsNaN(value.Float()) {
			err = ErrNaNValue
			break
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
		err = ErrMapNotSupported
	case reflect.Ptr:
		err = m.serializePointer(value, fieldMetadata, buffer)
	case reflect.Struct:
		err = m.serializeStructure(value, fieldMetadata, buffer)
	case reflect.Interface:
		err = m.serializeInterface(value, fieldMetadata, buffer)
	}
	return err
}

func (m *Serializer) serializeInterface(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (err error) {
	// write first byte only if AllowNil set to true
	if isNil, err := writeNilFlag(value, fieldMetadata, buffer); err != nil {
		return fmt.Errorf("%w: interface cannot have nil value", err)
	} else if isNil {
		return nil
	}

	interfaceType := reflect.TypeOf(value.Interface())
	encodedType, err := m.EncodeType(interfaceType)
	if err != nil {
		return
	}
	buffer.WriteUint32(encodedType)
	return m.SerializeValue(value.Elem(), fieldMetadata, buffer)
}

func (m *Serializer) serializeStructure(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (err error) {
	valueType := value.Type()
	// individual serialization for time.Time
	if valueType == reflect.TypeOf(time.Time{}) {
		buffer.WriteTime(value.Interface().(time.Time))
		return
	}

	if processed, err := m.tryBuiltInSerializer(value, valueType, fieldMetadata, buffer); err != nil || processed {
		// serialize using built-in serializer if it's available
		return err
	}

	// serialize as regular struct
	return m.serializeFields(value, buffer)
}

func (m *Serializer) serializeFields(value reflect.Value, buffer *marshalutil.MarshalUtil) (err error) {
	structType := value.Type()
	serializedFields, err := m.fieldCache.Fields(structType)
	if err != nil {
		return err
	}
	for _, fieldMeta := range serializedFields {
		fieldValue := value.Field(fieldMeta.Idx)

		if err = m.SerializeValue(fieldValue, fieldMeta, buffer); err != nil {
			err = fmt.Errorf("%w: error while serializing field: '%s'", err, fieldMeta.Name)
			break
		}
	}
	return
}

func (m *Serializer) serializePointer(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (err error) {
	// write first byte only if AllowNil set to true
	if isNil, err := writeNilFlag(value, fieldMetadata, buffer); err != nil {
		return fmt.Errorf("%w: pointer cannot have nil value", err)
	} else if isNil {
		return nil
	}

	// if pointer implements built-in serialization, use it
	valueType := reflect.TypeOf(value.Interface())
	if processed, err := m.tryBuiltInSerializer(value, valueType, fieldMetadata, buffer); err != nil || processed {
		return err
	}

	// or else serialize as regular pointer
	return m.SerializeValue(value.Elem(), fieldMetadata, buffer)
}

func (m *Serializer) serializeSlice(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) error {
	if value.Type().Elem().Kind() == reflect.Uint8 {
		// serialize slice of bytes individually to optimize execution
		return m.serializeByteSlice(value, fieldMetadata, buffer)
	} else if fieldMetadata.LexicalOrder || fieldMetadata.NoDuplicates {
		// if lexical order or no duplicates is required, perform necessary preprocessing to sort or remove duplicates
		return m.serializeConstrainedSlice(value, fieldMetadata, buffer)
	} else {
		// validate slice length according to specified values
		return m.serializeRegularSlice(value, fieldMetadata, buffer)
	}
}

func (m *Serializer) serializeRegularSlice(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (err error) {
	if err = ValidateLength(value.Len(), fieldMetadata.MinSliceLength, fieldMetadata.MaxSliceLength); err != nil {
		return
	}

	// write slice length and its elems
	if err = WriteLen(value.Len(), fieldMetadata.LengthPrefixType, buffer); err != nil {
		return fmt.Errorf("%w: error serializing slice length", err)
	}

	for i := 0; i < value.Len(); i++ {
		if err = m.SerializeValue(value.Index(i), fieldMetadata, buffer); err != nil {
			return fmt.Errorf("%w: error while serializing slice element on index %d", err, i)
		}
	}
	return
}

func (m *Serializer) serializeConstrainedSlice(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (err error) {
	elems := make([][]byte, 0)
	elemsMap := make(map[string]customtypes.Empty)

	for i := 0; i < value.Len(); i++ {
		// serialize slice element
		elemBuffer := marshalutil.New()
		if err = m.SerializeValue(value.Index(i), fieldMetadata, elemBuffer); err != nil {
			return fmt.Errorf("%w: error serializing slice value on index %d", err, i)
		}
		elemBytes := elemBuffer.Bytes()

		// check if no duplicate elements are serialized, skip any duplicates
		if fieldMetadata.NoDuplicates {
			elemBytesString := typeutils.BytesToString(elemBytes)
			if _, seenElem := elemsMap[elemBytesString]; fieldMetadata.NoDuplicates && seenElem {
				return fmt.Errorf("%w: index %d", ErrNoDuplicatesViolated, i)
			}
			elemsMap[elemBytesString] = customtypes.Void
		}

		// check if lexical order is preserved, return error otherwise
		if fieldMetadata.LexicalOrder && i > 0 && bytes.Compare(elems[i-1], elemBytes) >= 0 {
			return fmt.Errorf("%w: index %d", ErrLexicalOrderViolated, i)
		}

		elems = append(elems, elemBytes)
	}

	// validate length of the slice after preprocessing
	if err = ValidateLength(len(elems), fieldMetadata.MinSliceLength, fieldMetadata.MaxSliceLength); err != nil {
		return err
	}

	// write new length of the slice and its elems to the buffer
	if err = WriteLen(len(elems), fieldMetadata.LengthPrefixType, buffer); err != nil {
		return fmt.Errorf("%w: error serializing slice length", err)
	}

	for _, sortedElem := range elems {
		buffer.WriteBytes(sortedElem)
	}
	return
}

func (m *Serializer) serializeByteSlice(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (err error) {
	if err = ValidateLength(value.Len(), fieldMetadata.MinSliceLength, fieldMetadata.MaxSliceLength); err != nil {
		return
	}

	// write new length of the slice and its elems to the buffer
	if err = WriteLen(value.Len(), fieldMetadata.LengthPrefixType, buffer); err != nil {
		return fmt.Errorf("%w: error serializing slice length", err)
	}
	buffer.WriteBytes(value.Bytes())
	return
}

func (m *Serializer) serializeArray(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (err error) {
	arrayLen := value.Len()
	if value.Type().Elem().Kind() == reflect.Uint8 {
		// individual serialization for byte arrays in order to optimize execution
		for i := 0; i < arrayLen; i++ {
			buffer.WriteByte(uint8(value.Index(i).Uint()))
		}
		return
	}

	for i := 0; i < arrayLen; i++ {
		if err = m.SerializeValue(value.Index(i), fieldMetadata, buffer); err != nil {
			break
		}
	}
	return
}

func (m *Serializer) tryBuiltInSerializer(value reflect.Value, valueType reflect.Type, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (processed bool, err error) {
	if valueType.Implements(binaryMarshallerType) {
		processed = true
		err = m.serializeBinaryMarshaller(value, fieldMetadata, buffer)
	} else if valueType.Implements(autoserializerBinarySerializerType) {
		processed = true
		err = m.serializeBinarySerializer(value, fieldMetadata, buffer)
	}
	return
}

func (m *Serializer) serializeBinarySerializer(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (err error) {
	// if the structure implements BinarySerializer
	// writing number of bytes, because if the structure has been serialized correctly,
	// it will deserialize correctly and use only as many bytes from the buffer as needed, leaving others untouched
	marshaller := value.Interface().(BinarySerializer)
	var bytesMarshalled []byte

	if bytesMarshalled, err = marshaller.SerializeBytes(m, fieldMetadata); err != nil {
		return
	}

	buffer.WriteBytes(bytesMarshalled)
	return
}

func (m *Serializer) serializeBinaryMarshaller(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (err error) {
	// if the struct implements encoding.BinaryUnmarshaller
	// serialize using built-in method
	marshaller := value.Interface().(encoding.BinaryMarshaler)
	var bytesMarshalled []byte

	if bytesMarshalled, err = marshaller.MarshalBinary(); err != nil {
		err = fmt.Errorf("%w: error while serializing structure", err)
		return
	}

	// write number of bytes the struct was serialized to
	if err = WriteLen(len(bytesMarshalled), fieldMetadata.LengthPrefixType, buffer); err != nil {
		err = fmt.Errorf("%w: error while serializing length of the serialized structure", err)
		return
	}
	// write byte slice of serialized structure
	buffer.WriteBytes(bytesMarshalled)
	return
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
		return 0, fmt.Errorf("%w: %d", ErrUnknownLengthPrefix, lenPrefixType)
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
		return fmt.Errorf("%w: %d", ErrUnknownLengthPrefix, lenPrefixType)
	}
	return nil
}

// ValidateLength is used to make sure that the length of a collection is within bounds specified in struct tags.
func ValidateLength(length int, minSliceLen int, maxSliceLen int) (err error) {
	if length < minSliceLen {
		err = fmt.Errorf("%w: min %d elements instead of %d", ErrSliceMinLength, minSliceLen, length)
		return
	}
	if maxSliceLen > 0 && length > maxSliceLen {
		err = fmt.Errorf("%w: max %d elements instead of %d", ErrSliceMaxLength, maxSliceLen, length)
		return
	}
	return
}

func checkIfNil(fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (bool, error) {
	// if the value can be nil, read first byte to check whether it has value
	if fieldMetadata.AllowNil {
		nilPointer, err := buffer.ReadByte()
		if err != nil {
			return false, fmt.Errorf("%w: error reading nil flag byte", err)
		}
		// if pointer prefix is 0 then set value of the pointer to nil
		isNil := nilPointer == 0
		return isNil, nil
	}
	return false, nil
}

func writeNilFlag(value reflect.Value, fieldMetadata FieldMetadata, buffer *marshalutil.MarshalUtil) (isNil bool, err error) {
	if value.IsNil() && fieldMetadata.AllowNil {
		isNil = true
		buffer.WriteByte(byte(0))
	} else if fieldMetadata.AllowNil {
		buffer.WriteByte(byte(1))
	} else if value.IsNil() && !fieldMetadata.AllowNil {
		err = ErrNilNotAllowed
	}
	return
}

package valuerange

import (
	"fmt"
	"strconv"

	"golang.org/x/xerrors"

	"github.com/izuc/zipp.foundation/serializer/v2/marshalutil"
)

// region ValueType ////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	// Int8ValueType represents the type of Int8Values.
	Int8ValueType ValueType = iota

	// Int16ValueType represents the type of Int16Values.
	Int16ValueType

	// Int32ValueType represents the type of Int32Values.
	Int32ValueType

	// Int64ValueType represents the type of Int64Values.
	Int64ValueType

	// Uint8ValueType represents the type of Uint8Values.
	Uint8ValueType

	// Uint16ValueType represents the type of Uint16ValueType.
	Uint16ValueType

	// Uint32ValueType represents the type of Uint32ValueType.
	Uint32ValueType

	// Uint64ValueType represents the type of Uint64Values.
	Uint64ValueType
)

// ValueTypeNames contains a dictionary of the names of ValueTypes.
var ValueTypeNames = [...]string{
	"Int8ValueType",
	"Int16ValueType",
	"Int32ValueType",
	"Int64ValueType",
	"Uint8ValueType",
	"Uint16ValueType",
	"Uint32ValueType",
	"Uint64ValueType",
}

// ValueType represents the type different kinds of Values.
type ValueType int8

// ValueTypeFromBytes unmarshals a ValueType from a sequence of bytes.
func ValueTypeFromBytes(valueTypeBytes []byte) (valueType ValueType, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(valueTypeBytes)
	if valueType, err = ValueTypeFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse ValueType from MarshalUtil: %w", err)

		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// ValueTypeFromMarshalUtil unmarshals a ValueType using a MarshalUtil (for easier unmarshalling).
func ValueTypeFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (valueType ValueType, err error) {
	valueTypeByte, err := marshalUtil.ReadByte()
	if err != nil {
		err = xerrors.Errorf("failed to parse ValueType (%v): %w", err, ErrParseBytesFailed)

		return
	}

	if valueType = ValueType(valueTypeByte); valueType > Uint64ValueType {
		err = xerrors.Errorf("unsupported ValueType (%X): %w", valueType, ErrParseBytesFailed)

		return
	}

	return
}

// Bytes returns a marshaled version of the ValueType.
func (v ValueType) Bytes() []byte {
	return []byte{byte(v)}
}

// String returns a human-readable representation of the ValueType.
func (v ValueType) String() string {
	if int(v) >= len(ValueTypeNames) {
		return fmt.Sprintf("ValueType(%X)", uint8(v))
	}

	return ValueTypeNames[v]
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Value ////////////////////////////////////////////////////////////////////////////////////////////////////////

// Value is an interface that is used by the ValueRanges to compare different Values. It is required to keep the
// ValueRange generic.
type Value interface {
	// Type returns the type of the Value. It can be used to tell different ValueTypes apart and write polymorphic code.
	Type() ValueType

	// Compare return 0 if the other Value is identical, -1 if it is bigger and 1 if it is smaller.
	Compare(other Value) int

	// Bytes returns a marshaled version of the Value.
	Bytes() []byte

	// String returns a human-readable version of the Value.
	String() string
}

// ValueFromBytes unmarshals a Value from a sequence of bytes.
func ValueFromBytes(valueBytes []byte) (value Value, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(valueBytes)
	if value, err = ValueFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Value from MarshalUtil: %w", err)

		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// ValueFromMarshalUtil unmarshals a Value using a MarshalUtil (for easier unmarshalling).
func ValueFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (Value, error) {
	valueType, err := ValueTypeFromMarshalUtil(marshalUtil)
	if err != nil {
		return nil, xerrors.Errorf("failed to parse ValueType from MarshalUtil: %w", err)
	}
	marshalUtil.ReadSeek(-1)

	var value Value
	switch valueType {
	case Int8ValueType:
		if value, err = Int8ValueFromMarshalUtil(marshalUtil); err != nil {
			return nil, xerrors.Errorf("failed to parse Int8Value: %w", err)
		}
	case Int16ValueType:
		if value, err = Int16ValueFromMarshalUtil(marshalUtil); err != nil {
			return nil, xerrors.Errorf("failed to parse Int16Value: %w", err)
		}
	case Int32ValueType:
		if value, err = Int32ValueFromMarshalUtil(marshalUtil); err != nil {
			return nil, xerrors.Errorf("failed to parse Int32Value: %w", err)
		}
	case Int64ValueType:
		if value, err = Int64ValueFromMarshalUtil(marshalUtil); err != nil {
			return nil, xerrors.Errorf("failed to parse Int64Value: %w", err)
		}
	case Uint8ValueType:
		if value, err = Uint8ValueFromMarshalUtil(marshalUtil); err != nil {
			return nil, xerrors.Errorf("failed to parse Uint8Value: %w", err)
		}
	case Uint16ValueType:
		if value, err = Uint16ValueFromMarshalUtil(marshalUtil); err != nil {
			return nil, xerrors.Errorf("failed to parse Uint16Value: %w", err)
		}
	case Uint32ValueType:
		if value, err = Uint32ValueFromMarshalUtil(marshalUtil); err != nil {
			return nil, xerrors.Errorf("failed to parse Uint32Value: %w", err)
		}
	case Uint64ValueType:
		if value, err = Uint64ValueFromMarshalUtil(marshalUtil); err != nil {
			return nil, xerrors.Errorf("failed to parse Uint64Value: %w", err)
		}
	default:
		return nil, xerrors.Errorf("unsupported ValueType (%X): %w", valueType, ErrParseBytesFailed)
	}

	return value, nil
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Int8Value ////////////////////////////////////////////////////////////////////////////////////////////////////

// Int8Value is a wrapper for int8 values that makes these values compatible with the Value interface so they can be
// used in ValueRanges.
type Int8Value int8

// Int8ValueFromBytes unmarshals a Int8Value from a sequence of bytes.
func Int8ValueFromBytes(bytes []byte) (int8Value Int8Value, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if int8Value, err = Int8ValueFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Int8Value from MarshalUtil: %w", err)

		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// Int8ValueFromMarshalUtil unmarshals an Int8Value using a MarshalUtil (for easier unmarshalling).
func Int8ValueFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (int8Value Int8Value, err error) {
	valueType, err := ValueTypeFromMarshalUtil(marshalUtil)
	if err != nil {
		err = xerrors.Errorf("failed to parse ValueType from MarshalUtil: %w", err)

		return
	}
	if valueType != Int8ValueType {
		err = xerrors.Errorf("invalid ValueType (%s): %w", valueType, ErrParseBytesFailed)

		return
	}

	value, err := marshalUtil.ReadInt8()
	if err != nil {
		err = xerrors.Errorf("failed to read int8 (%v): %w", err, ErrParseBytesFailed)

		return
	}
	int8Value = Int8Value(value)

	return
}

// Type returns the type of the Value. It can be used to tell different ValueTypes apart and write polymorphic code.
func (i Int8Value) Type() ValueType {
	return Int8ValueType
}

// Compare return 0 if the other Value is identical, -1 if it is bigger and 1 if it is smaller.
func (i Int8Value) Compare(other Value) int {
	typeCastedOtherValue, typeCastOK := other.(Int8Value)
	if !typeCastOK {
		panic("can only compare an Int8Value to another Int8Value")
	}

	if typeCastedOtherValue == i {
		return 0
	}

	if typeCastedOtherValue > i {
		return -1
	}

	return 1
}

// Bytes returns a marshaled version of the Value.
func (i Int8Value) Bytes() []byte {
	return marshalutil.New(1 + marshalutil.Int8Size).
		Write(Int8ValueType).
		WriteInt8(int8(i)).
		Bytes()
}

// String returns a human-readable version of the Value.
func (i Int8Value) String() string {
	return "Int8Value(" + strconv.FormatInt(int64(i), 10) + ")"
}

// code contract (make sure the type implements all required methods).
var _ Value = Int8Value(0)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Int16Value ///////////////////////////////////////////////////////////////////////////////////////////////////

// Int16Value is a wrapper for int16 values that makes these values compatible with the Value interface so they can be
// used in ValueRanges.
type Int16Value int16

// Int16ValueFromBytes unmarshals a Int16Value from a sequence of bytes.
func Int16ValueFromBytes(bytes []byte) (int16Value Int16Value, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if int16Value, err = Int16ValueFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Int16Value from MarshalUtil: %w", err)

		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// Int16ValueFromMarshalUtil unmarshals an Int16Value using a MarshalUtil (for easier unmarshalling).
func Int16ValueFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (int16Value Int16Value, err error) {
	valueType, err := ValueTypeFromMarshalUtil(marshalUtil)
	if err != nil {
		err = xerrors.Errorf("failed to parse ValueType from MarshalUtil: %w", err)

		return
	}
	if valueType != Int16ValueType {
		err = xerrors.Errorf("invalid ValueType (%s): %w", valueType, ErrParseBytesFailed)

		return
	}

	value, err := marshalUtil.ReadInt16()
	if err != nil {
		err = xerrors.Errorf("failed to read int16 (%v): %w", err, ErrParseBytesFailed)

		return
	}
	int16Value = Int16Value(value)

	return
}

// Type returns the type of the Value. It can be used to tell different ValueTypes apart and write polymorphic code.
func (i Int16Value) Type() ValueType {
	return Int16ValueType
}

// Compare return 0 if the other Value is identical, -1 if it is bigger and 1 if it is smaller.
func (i Int16Value) Compare(other Value) int {
	typeCastedOtherValue, typeCastOK := other.(Int16Value)
	if !typeCastOK {
		panic("can only compare an Int16Value to another Int16Value")
	}

	if typeCastedOtherValue == i {
		return 0
	}

	if typeCastedOtherValue > i {
		return -1
	}

	return 1
}

// Bytes returns a marshaled version of the Value.
func (i Int16Value) Bytes() []byte {
	return marshalutil.New(1 + marshalutil.Int16Size).
		Write(Int16ValueType).
		WriteInt16(int16(i)).
		Bytes()
}

// String returns a human-readable version of the Value.
func (i Int16Value) String() string {
	return "Int16Value(" + strconv.FormatInt(int64(i), 10) + ")"
}

// code contract (make sure the type implements all required methods).
var _ Value = Int16Value(0)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Int32Value ///////////////////////////////////////////////////////////////////////////////////////////////////

// Int32Value is a wrapper for int32 values that makes these values compatible with the Value interface so they can be
// used in ValueRanges.
type Int32Value int32

// Int32ValueFromBytes unmarshals a Int32Value from a sequence of bytes.
func Int32ValueFromBytes(bytes []byte) (int32Value Int32Value, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if int32Value, err = Int32ValueFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Int32Value from MarshalUtil: %w", err)

		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// Int32ValueFromMarshalUtil unmarshals an Int32Value using a MarshalUtil (for easier unmarshalling).
func Int32ValueFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (int32Value Int32Value, err error) {
	valueType, err := ValueTypeFromMarshalUtil(marshalUtil)
	if err != nil {
		err = xerrors.Errorf("failed to parse ValueType from MarshalUtil: %w", err)

		return
	}
	if valueType != Int32ValueType {
		err = xerrors.Errorf("invalid ValueType (%s): %w", valueType, ErrParseBytesFailed)

		return
	}

	value, err := marshalUtil.ReadInt32()
	if err != nil {
		err = xerrors.Errorf("failed to read int32 (%v): %w", err, ErrParseBytesFailed)

		return
	}
	int32Value = Int32Value(value)

	return
}

// Type returns the type of the Value. It can be used to tell different ValueTypes apart and write polymorphic code.
func (i Int32Value) Type() ValueType {
	return Int32ValueType
}

// Compare return 0 if the other Value is identical, -1 if it is bigger and 1 if it is smaller.
func (i Int32Value) Compare(other Value) int {
	typeCastedOtherValue, typeCastOK := other.(Int32Value)
	if !typeCastOK {
		panic("can only compare an Int32Value to another Int32Value")
	}

	if typeCastedOtherValue == i {
		return 0
	}

	if typeCastedOtherValue > i {
		return -1
	}

	return 1
}

// Bytes returns a marshaled version of the Value.
func (i Int32Value) Bytes() []byte {
	return marshalutil.New(1 + marshalutil.Int32Size).
		Write(Int32ValueType).
		WriteInt32(int32(i)).
		Bytes()
}

// String returns a human-readable version of the Value.
func (i Int32Value) String() string {
	return "Int32Value(" + strconv.FormatInt(int64(i), 10) + ")"
}

// code contract (make sure the type implements all required methods).
var _ Value = Int32Value(0)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Int64Value ///////////////////////////////////////////////////////////////////////////////////////////////////

// Int64Value is a wrapper for int64 values that makes these values compatible with the Value interface, so they can be
// used in ValueRanges.
type Int64Value int64

// Int64ValueFromBytes unmarshals a Int64Value from a sequence of bytes.
func Int64ValueFromBytes(bytes []byte) (int64Value Int64Value, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if int64Value, err = Int64ValueFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Int64Value from MarshalUtil: %w", err)

		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// Int64ValueFromMarshalUtil unmarshals an Int64Value using a MarshalUtil (for easier unmarshalling).
func Int64ValueFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (int64Value Int64Value, err error) {
	valueType, err := ValueTypeFromMarshalUtil(marshalUtil)
	if err != nil {
		err = xerrors.Errorf("failed to parse ValueType from MarshalUtil: %w", err)

		return
	}
	if valueType != Int64ValueType {
		err = xerrors.Errorf("invalid ValueType (%s): %w", valueType, ErrParseBytesFailed)

		return
	}

	value, err := marshalUtil.ReadInt64()
	if err != nil {
		err = xerrors.Errorf("failed to read int64 (%v): %w", err, ErrParseBytesFailed)

		return
	}
	int64Value = Int64Value(value)

	return
}

// Type returns the type of the Value. It can be used to tell different ValueTypes apart and write polymorphic code.
func (i Int64Value) Type() ValueType {
	return Int64ValueType
}

// Compare return 0 if the other Value is identical, -1 if it is bigger and 1 if it is smaller.
func (i Int64Value) Compare(other Value) int {
	typeCastedOtherValue, typeCastOK := other.(Int64Value)
	if !typeCastOK {
		panic("can only compare an Int64Value to another Int64Value")
	}

	if typeCastedOtherValue == i {
		return 0
	}

	if typeCastedOtherValue > i {
		return -1
	}

	return 1
}

// Bytes returns a marshaled version of the Value.
func (i Int64Value) Bytes() []byte {
	return marshalutil.New(1 + marshalutil.Int64Size).
		Write(Int64ValueType).
		WriteInt64(int64(i)).
		Bytes()
}

// String returns a human-readable version of the Value.
func (i Int64Value) String() string {
	return "Int64Value(" + strconv.FormatInt(int64(i), 10) + ")"
}

// code contract (make sure the type implements all required methods).
var _ Value = Int64Value(0)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Uint8Value ///////////////////////////////////////////////////////////////////////////////////////////////////

// Uint8Value is a wrapper for uint8 values that makes these values compatible with the Value interface so they can be
// used in ValueRanges.
type Uint8Value uint8

// Uint8ValueFromBytes unmarshals a Uint8Value from a sequence of bytes.
func Uint8ValueFromBytes(bytes []byte) (uint8Value Uint8Value, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if uint8Value, err = Uint8ValueFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Uint8Value from MarshalUtil: %w", err)

		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// Uint8ValueFromMarshalUtil unmarshals an Uint8Value using a MarshalUtil (for easier unmarshalling).
func Uint8ValueFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (uint8Value Uint8Value, err error) {
	valueType, err := ValueTypeFromMarshalUtil(marshalUtil)
	if err != nil {
		err = xerrors.Errorf("failed to parse ValueType from MarshalUtil: %w", err)

		return
	}
	if valueType != Uint8ValueType {
		err = xerrors.Errorf("invalid ValueType (%s): %w", valueType, ErrParseBytesFailed)

		return
	}

	value, err := marshalUtil.ReadUint8()
	if err != nil {
		err = xerrors.Errorf("failed to read uint8 (%v): %w", err, ErrParseBytesFailed)

		return
	}
	uint8Value = Uint8Value(value)

	return
}

// Type returns the type of the Value. It can be used to tell different ValueTypes apart and write polymorphic code.
func (i Uint8Value) Type() ValueType {
	return Uint8ValueType
}

// Compare return 0 if the other Value is identical, -1 if it is bigger and 1 if it is smaller.
func (i Uint8Value) Compare(other Value) int {
	typeCastedOtherValue, typeCastOK := other.(Uint8Value)
	if !typeCastOK {
		panic("can only compare an Uint8Value to another Uint8Value")
	}

	if typeCastedOtherValue == i {
		return 0
	}

	if typeCastedOtherValue > i {
		return -1
	}

	return 1
}

// Bytes returns a marshaled version of the Value.
func (i Uint8Value) Bytes() []byte {
	return marshalutil.New(1 + marshalutil.Uint8Size).
		Write(Uint8ValueType).
		WriteUint8(uint8(i)).
		Bytes()
}

// String returns a human-readable version of the Value.
func (i Uint8Value) String() string {
	return "Uint8Value(" + strconv.FormatUint(uint64(i), 10) + ")"
}

// code contract (make sure the type implements all required methods).
var _ Value = Uint8Value(0)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Uint16Value //////////////////////////////////////////////////////////////////////////////////////////////////

// Uint16Value is a wrapper for uint16 values that makes these values compatible with the Value interface so they can be
// used in ValueRanges.
type Uint16Value uint16

// Uint16ValueFromBytes unmarshals a Uint16Value from a sequence of bytes.
func Uint16ValueFromBytes(bytes []byte) (uint16Value Uint16Value, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if uint16Value, err = Uint16ValueFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Uint16Value from MarshalUtil: %w", err)

		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// Uint16ValueFromMarshalUtil unmarshals an Uint16Value using a MarshalUtil (for easier unmarshalling).
func Uint16ValueFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (uint16Value Uint16Value, err error) {
	valueType, err := ValueTypeFromMarshalUtil(marshalUtil)
	if err != nil {
		err = xerrors.Errorf("failed to parse ValueType from MarshalUtil: %w", err)

		return
	}
	if valueType != Uint16ValueType {
		err = xerrors.Errorf("invalid ValueType (%s): %w", valueType, ErrParseBytesFailed)

		return
	}

	value, err := marshalUtil.ReadUint16()
	if err != nil {
		err = xerrors.Errorf("failed to read uint16 (%v): %w", err, ErrParseBytesFailed)

		return
	}
	uint16Value = Uint16Value(value)

	return
}

// Type returns the type of the Value. It can be used to tell different ValueTypes apart and write polymorphic code.
func (i Uint16Value) Type() ValueType {
	return Uint16ValueType
}

// Compare return 0 if the other Value is identical, -1 if it is bigger and 1 if it is smaller.
func (i Uint16Value) Compare(other Value) int {
	typeCastedOtherValue, typeCastOK := other.(Uint16Value)
	if !typeCastOK {
		panic("can only compare an Uint16Value to another Uint16Value")
	}

	if typeCastedOtherValue == i {
		return 0
	}

	if typeCastedOtherValue > i {
		return -1
	}

	return 1
}

// Bytes returns a marshaled version of the Value.
func (i Uint16Value) Bytes() []byte {
	return marshalutil.New(1 + marshalutil.Uint16Size).
		Write(Uint16ValueType).
		WriteUint16(uint16(i)).
		Bytes()
}

// String returns a human-readable version of the Value.
func (i Uint16Value) String() string {
	return "Uint16Value(" + strconv.FormatUint(uint64(i), 10) + ")"
}

// code contract (make sure the type implements all required methods).
var _ Value = Uint16Value(0)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Uint32Value //////////////////////////////////////////////////////////////////////////////////////////////////

// Uint32Value is a wrapper for uint32 values that makes these values compatible with the Value interface so they can be
// used in ValueRanges.
type Uint32Value uint32

// Uint32ValueFromBytes unmarshals a Uint32Value from a sequence of bytes.
func Uint32ValueFromBytes(bytes []byte) (uint32Value Uint32Value, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if uint32Value, err = Uint32ValueFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Uint32Value from MarshalUtil: %w", err)

		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// Uint32ValueFromMarshalUtil unmarshals an Uint32Value using a MarshalUtil (for easier unmarshalling).
func Uint32ValueFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (uint32Value Uint32Value, err error) {
	valueType, err := ValueTypeFromMarshalUtil(marshalUtil)
	if err != nil {
		err = xerrors.Errorf("failed to parse ValueType from MarshalUtil: %w", err)

		return
	}
	if valueType != Uint32ValueType {
		err = xerrors.Errorf("invalid ValueType (%s): %w", valueType, ErrParseBytesFailed)

		return
	}

	value, err := marshalUtil.ReadUint32()
	if err != nil {
		err = xerrors.Errorf("failed to read uint32 (%v): %w", err, ErrParseBytesFailed)

		return
	}
	uint32Value = Uint32Value(value)

	return
}

// Type returns the type of the Value. It can be used to tell different ValueTypes apart and write polymorphic code.
func (i Uint32Value) Type() ValueType {
	return Uint32ValueType
}

// Compare return 0 if the other Value is identical, -1 if it is bigger and 1 if it is smaller.
func (i Uint32Value) Compare(other Value) int {
	typeCastedOtherValue, typeCastOK := other.(Uint32Value)
	if !typeCastOK {
		panic("can only compare an Uint32Value to another Uint32Value")
	}

	if typeCastedOtherValue == i {
		return 0
	}

	if typeCastedOtherValue > i {
		return -1
	}

	return 1
}

// Bytes returns a marshaled version of the Value.
func (i Uint32Value) Bytes() []byte {
	return marshalutil.New(1 + marshalutil.Uint32Size).
		Write(Uint32ValueType).
		WriteUint32(uint32(i)).
		Bytes()
}

// String returns a human-readable version of the Value.
func (i Uint32Value) String() string {
	return "Uint32Value(" + strconv.FormatUint(uint64(i), 10) + ")"
}

// code contract (make sure the type implements all required methods).
var _ Value = Uint32Value(0)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Uint64Value ///////////////////////////////////////////////////////////////////////////////////////////////////

// Uint64Value is a wrapper for uint64 values that makes these values compatible with the Value interface so they can be
// used in ValueRanges.
type Uint64Value uint64

// Uint64ValueFromBytes unmarshals a Uint64Value from a sequence of bytes.
func Uint64ValueFromBytes(bytes []byte) (uint64Value Uint64Value, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if uint64Value, err = Uint64ValueFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Uint64Value from MarshalUtil: %w", err)

		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// Uint64ValueFromMarshalUtil unmarshals an Uint64Value using a MarshalUtil (for easier unmarshalling).
func Uint64ValueFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (uint64Value Uint64Value, err error) {
	valueType, err := ValueTypeFromMarshalUtil(marshalUtil)
	if err != nil {
		err = xerrors.Errorf("failed to parse ValueType from MarshalUtil: %w", err)

		return
	}
	if valueType != Uint64ValueType {
		err = xerrors.Errorf("invalid ValueType (%s): %w", valueType, ErrParseBytesFailed)

		return
	}

	value, err := marshalUtil.ReadUint64()
	if err != nil {
		err = xerrors.Errorf("failed to read uint64 (%v): %w", err, ErrParseBytesFailed)

		return
	}
	uint64Value = Uint64Value(value)

	return
}

// Type returns the type of the Value. It can be used to tell different ValueTypes apart and write polymorphic code.
func (i Uint64Value) Type() ValueType {
	return Uint64ValueType
}

// Compare return 0 if the other Value is identical, -1 if it is bigger and 1 if it is smaller.
func (i Uint64Value) Compare(other Value) int {
	typeCastedOtherValue, typeCastOK := other.(Uint64Value)
	if !typeCastOK {
		panic("can only compare an Uint64Value to another Uint64Value")
	}

	if typeCastedOtherValue == i {
		return 0
	}

	if typeCastedOtherValue > i {
		return -1
	}

	return 1
}

// Bytes returns a marshaled version of the Value.
func (i Uint64Value) Bytes() []byte {
	return marshalutil.New(1 + marshalutil.Uint64Size).
		Write(Uint64ValueType).
		WriteUint64(uint64(i)).
		Bytes()
}

// String returns a human-readable version of the Value.
func (i Uint64Value) String() string {
	return "Uint64Value(" + strconv.FormatUint(uint64(i), 10) + ")"
}

// code contract (make sure the type implements all required methods).
var _ Value = Uint64Value(0)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

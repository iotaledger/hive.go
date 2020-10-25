package valuerange

import (
	"strconv"

	"github.com/iotaledger/hive.go/cerrors"
	"github.com/iotaledger/hive.go/marshalutil"
	"golang.org/x/xerrors"
)

// region Value ////////////////////////////////////////////////////////////////////////////////////////////////////////

// Value is an interface that is used by the ValueRanges to compare different Values. It is required to keep the
// ValueRange generic.
type Value interface {
	// Compare return 0 if the other Value is identical, -1 if it is bigger and 1 if it is smaller.
	Compare(other Value) int

	// Bytes returns a marshaled version of the Value.
	Bytes() []byte

	// String returns a human readable version of the Value.
	String() string
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Int64Value ///////////////////////////////////////////////////////////////////////////////////////////////////

// Int64Value is a wrapper for int64 values that makes these values compatible with the Value interface so they can be
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

// Int64ValueFromMarshalUtil unmarshals a BLSSignature using a MarshalUtil (for easier unmarshaling).
func Int64ValueFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (int64Value Int64Value, err error) {
	value, err := marshalUtil.ReadInt64()
	if err != nil {
		err = xerrors.Errorf("failed to read int64 (%v): %w", err, cerrors.ErrParseBytesFailed)
		return
	}
	int64Value = Int64Value(value)

	return
}

// Compare return 0 if the other Value is identical, -1 if it is bigger and 1 if it is smaller.
func (i Int64Value) Compare(other Value) int {
	typeCastedOtherValue, typeCastOK := other.(Int64Value)
	if !typeCastOK {
		panic("can only compare Int64Values to other Int64Values")
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
	return marshalutil.New(marshalutil.INT64_SIZE).WriteInt64(int64(i)).Bytes()
}

// String returns a human readable version of the Value.
func (i Int64Value) String() string {
	return strconv.FormatInt(int64(i), 10)
}

// code contract (make sure the type implements all required methods)
var _ Value = Int64Value(0)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

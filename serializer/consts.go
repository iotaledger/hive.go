package serializer

import "math"

const (
	// OneByte is the byte size of a single byte.
	OneByte = 1
	// Int16ByteSize is the byte size of an int16.
	Int16ByteSize = 2
	// UInt16ByteSize is the byte size of a uint16.
	UInt16ByteSize = 2
	// Int32ByteSize is the byte size of an int32.
	Int32ByteSize = 4
	// UInt32ByteSize is the byte size of a uint32.
	UInt32ByteSize = 4
	// Float32ByteSize is the byte size of a float32.
	Float32ByteSize = 4
	// Int64ByteSize is the byte size of an int64.
	Int64ByteSize = 8
	// UInt64ByteSize is the byte size of an uint64.
	UInt64ByteSize = 8
	// UInt256ByteSize is the byte size of an uint256.
	UInt256ByteSize = 32
	// Float64ByteSize is the byte size of a float64.
	Float64ByteSize = 8
	// TypeDenotationByteSize is the size of a type denotation.
	TypeDenotationByteSize = UInt32ByteSize
	// SmallTypeDenotationByteSize is the size of a type denotation for a small range of possible values.
	SmallTypeDenotationByteSize = OneByte
	// PayloadLengthByteSize is the size of the payload length denoting bytes.
	PayloadLengthByteSize = UInt32ByteSize
	// MinPayloadByteSize is the minimum size of a payload (together with its length denotation).
	MinPayloadByteSize = UInt32ByteSize + OneByte
	// MaxNanoTimestampInt64Seconds is the maximum number of seconds that fit into a nanosecond-precision int64 timestamp.
	MaxNanoTimestampInt64Seconds = math.MaxInt64 / 1_000_000_000
)

// TypeDenotationType defines a type denotation.
type TypeDenotationType byte

//go:generate stringer -type=TypeDenotationType

const (
	// TypeDenotationUint32 defines a denotation which defines a type ID by a uint32.
	TypeDenotationUint32 TypeDenotationType = iota
	// TypeDenotationByte defines a denotation which defines a type ID by a byte.
	TypeDenotationByte
	// TypeDenotationNone defines that there is no type denotation.
	TypeDenotationNone
)

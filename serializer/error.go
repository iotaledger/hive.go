package serializer

import (
	"encoding/binary"

	"github.com/iotaledger/hive.go/ierrors"
)

var (
	// ErrInvalidBytes gets returned when data is invalid for deserialization.
	ErrInvalidBytes = ierrors.New("invalid bytes")
	// ErrDeserializationTypeMismatch gets returned when a denoted type for a given object is mismatched.
	// For example, while trying to deserialize a signature unlock block, a reference unlock block is seen.
	ErrDeserializationTypeMismatch = ierrors.New("data type is invalid for deserialization")
	// ErrUnknownArrayValidationMode gets returned for unknown array validation modes.
	ErrUnknownArrayValidationMode = ierrors.New("unknown array validation mode")
	// ErrArrayValidationMinElementsNotReached gets returned if the count of elements is too small.
	ErrArrayValidationMinElementsNotReached = ierrors.New("min count of elements within the array not reached")
	// ErrArrayValidationMaxElementsExceeded gets returned if the count of elements is too big.
	ErrArrayValidationMaxElementsExceeded = ierrors.New("max count of elements within the array exceeded")
	// ErrArrayValidationViolatesUniqueness gets returned if the array elements are not unique.
	ErrArrayValidationViolatesUniqueness = ierrors.New("array elements must be unique")
	// ErrArrayValidationViolatesTypeUniqueness gets returned if the array contains the same type multiple times.
	ErrArrayValidationViolatesTypeUniqueness = ierrors.New("array elements must be of a unique type")
	// ErrArrayValidationOrderViolatesLexicalOrder gets returned if the array elements are not in lexical order.
	ErrArrayValidationOrderViolatesLexicalOrder = ierrors.New("array elements must be in their lexical order (byte wise)")
	// ErrArrayValidationTypesNotOccurred gets returned if not all types as specified in an ArrayRules.MustOccur are in an array.
	ErrArrayValidationTypesNotOccurred = ierrors.New("not all needed types are present")
	// ErrDeserializationNotEnoughData gets returned if there is not enough data available to deserialize a given object.
	ErrDeserializationNotEnoughData = ierrors.New("not enough data for deserialization")
	// ErrDeserializationInvalidBoolValue gets returned when a bool value is tried to be read but it is neither 0 or 1.
	ErrDeserializationInvalidBoolValue = ierrors.New("invalid bool value")
	// ErrDeserializationLengthInvalid gets returned if a length denotation exceeds a specified limit.
	ErrDeserializationLengthInvalid = ierrors.New("length denotation invalid")
	// ErrDeserializationLengthMinNotReached gets returned if a length denotation is less than a specified limit.
	ErrDeserializationLengthMinNotReached = ierrors.Wrap(ErrDeserializationLengthInvalid, "min length not reached")
	// ErrDeserializationLengthMaxExceeded gets returned if a length denotation is more than a specified limit.
	ErrDeserializationLengthMaxExceeded = ierrors.Wrap(ErrDeserializationLengthInvalid, "max length exceeded")
	// ErrDeserializationNotAllConsumed gets returned if not all bytes were consumed during deserialization of a given type.
	ErrDeserializationNotAllConsumed = ierrors.New("not all data has been consumed but should have been")
	// ErrUint256NumNegative gets returned if a supposed uint256 has a sign bit.
	ErrUint256NumNegative = ierrors.New("uint256 is negative")
	// ErrSliceLengthTooShort gets returned if a slice is less than a min length.
	ErrSliceLengthTooShort = ierrors.New("slice length is too short")
	// ErrSliceLengthTooLong gets returned if a slice exceeds a max length.
	ErrSliceLengthTooLong = ierrors.New("slice length is too long")
	// ErrStringTooShort gets returned if a string is less than a min length.
	ErrStringTooShort = ierrors.New("string is too short")
	// ErrStringTooLong gets returned if a string exceeds a max length.
	ErrStringTooLong = ierrors.New("string is too long")
	// ErrUint256TooBig gets returned when a supposed uint256 big.Int value is more than 32 bytes in size.
	ErrUint256TooBig = ierrors.New("uint256 big int is too big")
	// ErrUint256Nil gets returned when a uint256 *big.Int is nil.
	ErrUint256Nil = ierrors.New("uint256 must not be nil")
)

// CheckType checks that the denoted type equals the shouldType.
func CheckType(data []byte, shouldType uint32) error {
	if len(data) < UInt32ByteSize {
		return ierrors.Wrap(ErrDeserializationNotEnoughData, "can't check type denotation")
	}
	actualType := binary.LittleEndian.Uint32(data)
	if actualType != shouldType {
		return ierrors.Wrapf(ErrDeserializationTypeMismatch, "type denotation must be %d but is %d", shouldType, actualType)
	}

	return nil
}

// CheckTypeByte checks that the denoted type byte equals the shouldType.
func CheckTypeByte(data []byte, shouldType byte) error {
	if len(data) == 0 {
		return ierrors.Wrap(ErrDeserializationNotEnoughData, "can't check type byte")
	}
	if data[0] != shouldType {
		return ierrors.Wrapf(ErrDeserializationTypeMismatch, "type denotation must be %d but is %d", shouldType, data[0])
	}

	return nil
}

// CheckExactByteLength checks that the given length equals exact.
func CheckExactByteLength(exact int, length int) error {
	if length != exact {
		return ierrors.Wrapf(ErrInvalidBytes, "data must be at exact %d bytes long but is %d", exact, length)
	}

	return nil
}

// CheckMinByteLength checks that length is at least min.
func CheckMinByteLength(min int, length int) error {
	if length < min {
		return ierrors.Wrapf(ErrDeserializationNotEnoughData, "data must be at least %d bytes long but is %d", min, length)
	}

	return nil
}

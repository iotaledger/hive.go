package serializer

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var (
	// ErrInvalidBytes gets returned when data is invalid for deserialization.
	ErrInvalidBytes = errors.New("invalid bytes")
	// ErrDeserializationTypeMismatch gets returned when a denoted type for a given object is mismatched.
	// For example, while trying to deserialize a signature unlock block, a reference unlock block is seen.
	ErrDeserializationTypeMismatch = errors.New("data type is invalid for deserialization")
	// ErrUnknownArrayValidationMode gets returned for unknown array validation modes.
	ErrUnknownArrayValidationMode = errors.New("unknown array validation mode")
	// ErrArrayValidationMinElementsNotReached gets returned if the count of elements is too small.
	ErrArrayValidationMinElementsNotReached = errors.New("min count of elements within the array not reached")
	// ErrArrayValidationMaxElementsExceeded gets returned if the count of elements is too big.
	ErrArrayValidationMaxElementsExceeded = errors.New("max count of elements within the array exceeded")
	// ErrArrayValidationViolatesUniqueness gets returned if the array elements are not unique.
	ErrArrayValidationViolatesUniqueness = errors.New("array elements must be unique")
	// ErrArrayValidationOrderViolatesLexicalOrder gets returned if the array elements are not in lexical order.
	ErrArrayValidationOrderViolatesLexicalOrder = errors.New("array elements must be in their lexical order (byte wise)")
	// ErrDeserializationNotEnoughData gets returned if there is not enough data available to deserialize a given object.
	ErrDeserializationNotEnoughData = errors.New("not enough data for deserialization")
	// ErrDeserializationInvalidBoolValue gets returned when a bool value is tried to be read but it is neither 0 or 1.
	ErrDeserializationInvalidBoolValue = errors.New("invalid bool value")
	// ErrDeserializationLengthInvalid gets returned if a length denotation exceeds a specified limit.
	ErrDeserializationLengthInvalid = errors.New("length denotation invalid")
	// ErrDeserializationNotAllConsumed gets returned if not all bytes were consumed during deserialization of a given type.
	ErrDeserializationNotAllConsumed = errors.New("not all data has been consumed but should have been")
)

// CheckType checks that the denoted type equals the shouldType.
func CheckType(data []byte, shouldType uint32) error {
	if len(data) < 4 {
		return fmt.Errorf("%w: can't check type denotation", ErrDeserializationNotEnoughData)
	}
	actualType := binary.LittleEndian.Uint32(data)
	if actualType != shouldType {
		return fmt.Errorf("%w: type denotation must be %d but is %d", ErrDeserializationTypeMismatch, shouldType, actualType)
	}
	return nil
}

// CheckTypeByte checks that the denoted type byte equals the shouldType.
func CheckTypeByte(data []byte, shouldType byte) error {
	if len(data) == 0 {
		return fmt.Errorf("%w: can't check type byte", ErrDeserializationNotEnoughData)
	}
	if data[0] != shouldType {
		return fmt.Errorf("%w: type denotation must be %d but is %d", ErrDeserializationTypeMismatch, shouldType, data[0])
	}
	return nil
}

// CheckExactByteLength checks that the given length equals exact.
func CheckExactByteLength(exact int, length int) error {
	if length != exact {
		return fmt.Errorf("%w: data must be at exact %d bytes long but is %d", ErrInvalidBytes, exact, length)
	}
	return nil
}

// CheckMinByteLength checks that length is min. min.
func CheckMinByteLength(min int, length int) error {
	if length < min {
		return fmt.Errorf("%w: data must be at least %d bytes long but is %d", ErrDeserializationNotEnoughData, min, length)
	}
	return nil
}

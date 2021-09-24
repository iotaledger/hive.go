package refseri

import "errors"

//ErrNotAllBytesRead error returned when not all bytes have been read from the buffer
var ErrNotAllBytesRead = errors.New("did not read all bytes from the buffer")

// ErrMapNotSupported error returned when trying to serialize/deserialize native map type
var ErrMapNotSupported = errors.New("native map type is not supported. use orderedmap instead")

//ErrDeserializeInterface error returned when there is a problem during interface serialization
var ErrDeserializeInterface = errors.New("couldn't deserialize interface")

//ErrUnpackAnonymous error returned when 'unpack' tag is added to anonymous field
var ErrUnpackAnonymous = errors.New("cannot unpack on non anonymous field")

//ErrUnpackNonStruct error returned when 'unpack' tag is added to a non-struct field
var ErrUnpackNonStruct = errors.New("cannot unpack on non struct field")

//ErrUnknownLengthPrefix error returned when length prefix struct tag is set to an unknown value
var ErrUnknownLengthPrefix = errors.New("unknown length prefix type")

//ErrNilNotAllowed error returned when non-optional pointer or interface field has nil value
var ErrNilNotAllowed = errors.New("nil value is not allowed")

//ErrNaNValue error returned when float type has NaN value
var ErrNaNValue = errors.New("NaN float value")

//ErrSliceMinLength error is returned when a slice does not have minimum required length
var ErrSliceMinLength = errors.New("collection is required to have min number of elements")

//ErrSliceMaxLength error is returned when a slice has more than maximum allowed length
var ErrSliceMaxLength = errors.New("collection is required to have max number of elements")

//ErrLexicalOrderViolated error is returned when slice is not sorted in byte lexical order
var ErrLexicalOrderViolated = errors.New("lexical order violated")

//ErrNoDuplicatesViolated error is returned when slice does not have unique elements
var ErrNoDuplicatesViolated = errors.New("no duplicates requirement violated")

// ErrUnexportedField error returned when trying to marshal unexported field
var ErrUnexportedField = errors.New("can't marshal un-exported field")

// ErrTypeNotRegistered error returned when trying to encode/decode a type that was not registered
var ErrTypeNotRegistered = errors.New("type not registered")

// ErrAlreadyRegistered error returned when trying to register a type multiple times
var ErrAlreadyRegistered = errors.New("type already registered")

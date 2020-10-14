package bls

import "errors"

var (
	// ErrBase58DecodeFailed is returned if a base58 encoded string can not be decoded.
	ErrBase58DecodeFailed = errors.New("failed to decode base58 encoded string")

	// ErrParseBytesFailed is returned if information can not be parsed from a sequence of bytes.
	ErrParseBytesFailed = errors.New("failed to parse bytes")

	// ErrBLSFailed is returned if any low level BLS method calls fail.
	ErrBLSFailed = errors.New("failed to execute BLS function")

	// ErrInvalidArgument is returned if a function gets called with an illegal argument.
	ErrInvalidArgument = errors.New("invalid argument")
)

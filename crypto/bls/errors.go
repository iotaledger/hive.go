package bls

import "github.com/iotaledger/hive.go/ierrors"

var (
	// ErrBase58DecodeFailed is returned if a base58 encoded string can not be decoded.
	ErrBase58DecodeFailed = ierrors.New("failed to decode base58 encoded string")

	// ErrParseBytesFailed is returned if information can not be parsed from a sequence of bytes.
	ErrParseBytesFailed = ierrors.New("failed to parse bytes")

	// ErrBLSFailed is returned if any low level BLS method calls fail.
	ErrBLSFailed = ierrors.New("failed to execute BLS function")

	// ErrInvalidArgument is returned if a function gets called with an illegal argument.
	ErrInvalidArgument = ierrors.New("invalid argument")
)

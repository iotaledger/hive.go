package types

import (
	"github.com/iotaledger/hive.go/cerrors"
	"github.com/iotaledger/hive.go/marshalutil"
	"golang.org/x/xerrors"
)

const (
	// False represents the equivalent of the boolean false value.
	False TriBool = iota

	// True represents the equivalent of the boolean true value.
	True

	// Maybe represents an indeterminate where we are not entirely sure if the value is True or False.
	Maybe
)

// TriBool represents a boolean value that can have an additional Maybe state.
type TriBool uint8

// TriBoolFromBytes unmarshals a TriBool from a sequence of bytes.
func TriBoolFromBytes(bytes []byte) (triBool TriBool, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if triBool, err = TriBoolFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse TriBool from MarshalUtil: %w", err)
		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// TriBoolFromMarshalUtil unmarshals a TriBool using a MarshalUtil (for easier unmarshaling).
func TriBoolFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (triBool TriBool, err error) {
	untypedTriBool, err := marshalUtil.ReadUint8()
	if err != nil {
		err = xerrors.Errorf("failed to parse TriBool (%v): %w", err, cerrors.ErrParseBytesFailed)
		return
	}
	if untypedTriBool >= 3 {
		err = xerrors.Errorf("failed to parse TriBool (out of bounds): %w", cerrors.ErrParseBytesFailed)
		return
	}
	triBool = TriBool(untypedTriBool)

	return
}

// Bytes returns a marshaled version of the TriBool.
func (t TriBool) Bytes() (marshaledTriBool []byte) {
	return []byte{byte(t)}
}

// String returns a human readable version of the TriBool.
func (t TriBool) String() (humanReadableTriBool string) {
	switch t {
	case 0:
		return "false"
	case 1:
		return "true"
	case 2:
		return "maybe"
	default:
		panic("invalid TriBool value")
	}
}

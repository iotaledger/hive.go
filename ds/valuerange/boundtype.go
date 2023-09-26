package valuerange

import (
	"fmt"

	"golang.org/x/xerrors"

	"github.com/izuc/zipp.foundation/serializer/v2/marshalutil"
)

// BoundType indicates whether an EndPoint of some ValueRange is contained in the ValueRange itself ("closed") or not
// ("open"). If a range is unbounded on a side, it is neither open nor closed on that side; the bound simply does not
// exist.
type BoundType uint8

const (
	// BoundTypeOpen indicates that the EndPoint value is considered part of the ValueRange ("inclusive").
	BoundTypeOpen BoundType = iota

	// BoundTypeClosed indicates that the EndPoint value is not considered part of the ValueRange ("exclusive").
	BoundTypeClosed
)

// BoundTypeNames contains a dictionary of the names of BoundTypes.
var BoundTypeNames = [...]string{
	"BoundTypeOpen",
	"BoundTypeClosed",
}

// BoundTypeFromBytes unmarshals a BoundType from a sequence of bytes.
func BoundTypeFromBytes(boundTypeBytes []byte) (boundType BoundType, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(boundTypeBytes)
	if boundType, err = BoundTypeFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse BoundType from MarshalUtil: %w", err)

		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// BoundTypeFromMarshalUtil unmarshals a BoundType using a MarshalUtil (for easier unmarshalling).
func BoundTypeFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (boundType BoundType, err error) {
	boundTypeByte, err := marshalUtil.ReadByte()
	if err != nil {
		err = xerrors.Errorf("failed to read BoundType (%v): %w", err, ErrParseBytesFailed)

		return
	}

	if boundType = BoundType(boundTypeByte); boundType > BoundTypeClosed {
		err = xerrors.Errorf("unsupported BoundType (%X): %w", boundType, ErrParseBytesFailed)

		return
	}

	return
}

// Bytes returns a marshaled version of the BoundType.
func (b BoundType) Bytes() []byte {
	return []byte{byte(b)}
}

// String returns a human-readable version of the BoundType.
func (b BoundType) String() string {
	if int(b) >= len(BoundTypeNames) {
		return fmt.Sprintf("BoundType(%X)", uint8(b))
	}

	return BoundTypeNames[b]
}

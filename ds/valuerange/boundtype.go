package valuerange

import (
	"fmt"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
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
		err = ierrors.Wrap(err, "failed to parse BoundType from MarshalUtil")

		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// BoundTypeFromMarshalUtil unmarshals a BoundType using a MarshalUtil (for easier unmarshalling).
func BoundTypeFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (boundType BoundType, err error) {
	boundTypeByte, err := marshalUtil.ReadByte()
	if err != nil {
		err = ierrors.Wrapf(ErrParseBytesFailed, "failed to read BoundType: %w", err)

		return
	}

	if boundType = BoundType(boundTypeByte); boundType > BoundTypeClosed {
		err = ierrors.Wrapf(ErrParseBytesFailed, "unsupported BoundType (%X)", boundType)

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

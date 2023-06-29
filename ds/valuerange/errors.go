package valuerange

import "github.com/iotaledger/hive.go/ierrors"

var (
	// ErrParseBytesFailed is returned if information can not be parsed from a sequence of bytes.
	ErrParseBytesFailed = ierrors.New("failed to parse bytes")
)

package typeutils

import (
	"encoding/binary"

	"github.com/iotaledger/hive.go/ierrors"
)

func Uint64FromBytes(bytes []byte) (object uint64, consumed int, err error) {
	if len(bytes) < 8 {
		return 0, 0, ierrors.New("not enough bytes to decode uint64")
	}

	return binary.LittleEndian.Uint64(bytes), 8, nil
}

func ByteArray32FromBytes(bytes []byte) ([32]byte, int, error) {
	if len(bytes) < 32 {
		return [32]byte{}, 0, ierrors.New("not enough bytes to decode [32]byte")
	}

	return [32]byte(bytes), 32, nil
}

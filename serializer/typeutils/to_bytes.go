package typeutils

import "encoding/binary"

func Uint64ToBytes(u uint64) ([]byte, error) {
	result := make([]byte, 8)
	binary.LittleEndian.PutUint64(result, u)

	return result, nil
}

func ByteArray32ToBytes(b [32]byte) ([]byte, error) {
	return b[:], nil
}

package stream

type allowedGenericTypes interface {
	~bool | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~int8 | ~int16 | ~int32 | ~int64 | ~[32]byte | ~[36]byte | ~[38]byte
}

package stream

type allowedGenericTypes interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | bool | ~[32]byte | ~[36]byte
}

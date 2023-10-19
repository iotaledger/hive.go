package types

type Empty struct{}

var Void = Empty{}

func (e Empty) Bytes() ([]byte, error) {
	return []byte{}, nil
}

func EmptyFromBytes([]byte) (object Empty, consumed int, err error) { return Empty{}, 0, nil }

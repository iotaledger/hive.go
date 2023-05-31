package types

type Empty struct{}

var Void = Empty{}

func (e Empty) Bytes() ([]byte, error) {
	return []byte{}, nil
}

func (e *Empty) FromBytes([]byte) (int, error) {
	return 0, nil
}

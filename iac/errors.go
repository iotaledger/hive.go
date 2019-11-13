package iac

import "github.com/pkg/errors"

var (
	ErrInvalidCharInput = errors.New("invalid character in input")
)

func NewErrDecodeFailed(cause error) *ErrDecodeFailed {
	return &ErrDecodeFailed{Inner: cause}
}

type ErrDecodeFailed struct {
	Inner error
}

func (e ErrDecodeFailed) Cause() error {
	return e.Inner
}

func (e ErrDecodeFailed) Error() string {
	return "decoding error: " + e.Error()
}

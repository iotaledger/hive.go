package dataflow

import (
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"
)

func Benchmark(b *testing.B) {
	x := func(param int, next func(int) error) error {
		return next(param + 1)
	}

	y := func(param int, next func(int) error) error {
		return next(param + 2)
	}

	z := func(param int, next func(int) error) error {
		return errors.Errorf("FAILED")
	}

	dataFlow1 := New(x, y, z)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		dataFlow1.Run(i)
	}
}

func Test(t *testing.T) {
	x := func(param int, next func(int) error) error {
		return next(param + 1)
	}

	y := func(param int, next func(int) error) error {
		return next(param + 2)
	}

	z := func(param int, next func(int) error) error {
		return errors.Errorf("FAILED")
	}

	result := 0
	assert.NoError(t, New(x, y, func(param int, next func(result int) error) error {
		result = param

		return next(param)
	}).Run(1))
	assert.NoError(t, New(x, y).Run(7))
	assert.Equal(t, 4, result)

	dataFlow2 := New(New(x, y).RunWithCallback, y, z)
	assert.Error(t, dataFlow2.Run(1))
}

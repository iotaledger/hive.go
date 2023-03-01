package dataflow

import (
	"fmt"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Benchmark(b *testing.B) {
	x := func(param int, next Next[int]) error {
		return next(param + 1)
	}

	y := func(param int, next Next[int]) error {
		return next(param + 2)
	}

	z := func(param int, next Next[int]) error {
		return errors.Errorf("FAILED")
	}

	dataFlow1 := New[int](x, y, z)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		require.NoError(b, dataFlow1.Run(i))
	}
}

func Test(t *testing.T) {
	x := func(param int, next Next[int]) error {
		return next(param + 1)
	}

	y := func(param int, next Next[int]) error {
		return next(param + 2)
	}

	z := func(param int, next Next[int]) error {
		return errors.Errorf("FAILED")
	}

	result := 0
	assert.NoError(t, New(x, y, func(param int, next Next[int]) error {
		result = param

		return next(param)
	}).Run(1))
	assert.NoError(t, New(x, y).Run(7))
	assert.Equal(t, 4, result)

	dataFlow2 := New(New(x, y).ChainedCommand, y, z)
	assert.Error(t, dataFlow2.Run(1))
}

func TestDataFlow_WithDoneCallback(t *testing.T) {
	x := func(param int, next Next[int]) error {
		fmt.Println("x")

		return next(param + 1)
	}

	y := func(param int, next Next[int]) error {
		fmt.Println("y")

		return next(param + 2)
	}

	z := func(param int, next Next[int]) error {
		return nil
	}

	dataFlow1 := New(x, y, z).WithSuccessCallback(func(param int) {
		fmt.Println("1 done")
	}).WithAbortCallback(func(param int) {
		fmt.Println("1 aborted")
	})

	dataFlow2 := New(dataFlow1.ChainedCommand, y).WithSuccessCallback(func(param int) {
		fmt.Println("2 done")
	}).WithAbortCallback(func(param int) {
		fmt.Println("2 aborted")
	})

	fmt.Println("START")
	require.NoError(t, dataFlow2.Run(1))
	fmt.Println("END")
}

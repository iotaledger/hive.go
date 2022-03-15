package dataflow

import (
	"fmt"
	"testing"

	"github.com/cockroachdb/errors"
)

func Test(t *testing.T) {
	x := func(param int, next func(int) error) error {
		fmt.Println("x START", param)
		defer fmt.Println("x END")

		return next(param + 1)
	}

	y := func(param int, next func(int) error) error {
		fmt.Println("y START", param)
		defer fmt.Println("y END")

		return next(param + 2)
	}

	z := func(param int, next func(int) error) error {
		fmt.Println("z START", param)
		defer fmt.Println("z END")

		return errors.Errorf("FAILED")
	}

	dataFlow1 := New(x, y)
	fmt.Println(dataFlow1.Run(1))
	fmt.Println(dataFlow1.Run(7))

	dataFlow2 := New(dataFlow1.ChainedCommand(), z)
	fmt.Println(dataFlow2.Run(1))
}

package dataflow

import (
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	step1 := func(params *int) (err error) {
		assert.Equal(t, 0, *params)

		*params++

		return
	}

	step2 := func(params *int) (err error) {
		assert.Equal(t, 1, *params)

		*params++

		return
	}

	step3 := func(params *int) (err error) {
		assert.Equal(t, 2, *params)

		return errors.Errorf("something went wrong")
	}

	{
		start := 0

		var successResult *int
		New(step1, step2).OnSuccess(func(params *int) {
			successResult = params
		}).OnError(func(err error, params *int) {
			t.Fail()
		}).Run(&start)

		assert.Equal(t, 2, *successResult)
	}

	{
		start := 0

		triggered := false
		New(step1, step2, step3).OnSuccess(func(params *int) {
			t.Fail()
		}).OnError(func(err error, params *int) {
			triggered = true

			assert.Error(t, err)
			assert.Equal(t, 2, *params)
		}).Run(&start)

		assert.True(t, triggered)
	}

}

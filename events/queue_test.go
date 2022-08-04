package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	executionResult := ""

	event1 := NewEvent(VoidCaller)
	event2 := NewEvent(VoidCaller)
	event3 := NewEvent(VoidCaller)

	event1.Hook(NewClosure(func() {
		executionResult += "first"
	}))
	event2.Hook(NewClosure(func() {
		executionResult += "second"
	}))
	event3.Hook(NewClosure(func() {
		executionResult += "third"
	}))

	q := NewQueue()
	q.Queue(event1)
	q.Queue(event2)

	event3.Trigger()
	q.Trigger()

	assert.Equal(t, executionResult, "thirdfirstsecond")
}

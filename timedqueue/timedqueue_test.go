package timedqueue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimedQueue(t *testing.T) {
	timedQ := New()

	// prepare a list to insert
	var elements []time.Time
	now := time.Now().Add(15 * time.Second)
	for i := 0; i < 10; i++ {
		elements = append(elements, now.Add(time.Duration(i)*time.Second))
	}

	// insert elements to timedQ
	var timeElements []*QueueElement
	for i := 0; i < 10; i++ {
		timeElements = append(timeElements, timedQ.Add(i, elements[i]))
	}
	assert.Equal(t, len(elements), timedQ.Size())

	// wait and Poll all elements, check the popped time is correct.
	for {
		topElement := timedQ.Poll(false).(int)
		popTime := time.Now()
		assert.True(t, (popTime.Sub(elements[topElement]) < time.Duration(1*time.Millisecond)))

		if topElement == len(elements)-1 {
			break
		}
	}

	assert.Equal(t, 0, timedQ.Size())
	timedQ.Shutdown()
}

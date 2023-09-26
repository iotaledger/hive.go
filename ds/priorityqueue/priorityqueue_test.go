package priorityqueue_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/izuc/zipp.foundation/ds/priorityqueue"
)

func TestPriorityQueue(t *testing.T) {
	queue := priorityqueue.New[queueElement]()
	queue.Push(2)
	queue.Push(3)
	peekedElement, exists := queue.Peek()
	assert.True(t, exists)
	assert.Equal(t, queueElement(2), peekedElement, "wrong element")
	poppedElement, exists := queue.Pop()
	assert.True(t, exists)
	assert.Equal(t, queueElement(2), poppedElement, "wrong element")
	peekedElement, exists = queue.Peek()
	assert.True(t, exists)
	assert.Equal(t, queueElement(3), peekedElement, "wrong element")
	queue.Push(1)
	peekedElement, exists = queue.Peek()
	assert.True(t, exists)
	assert.Equal(t, queueElement(1), peekedElement, "wrong element")
}

type queueElement int

func (q queueElement) Compare(other queueElement) int {
	if q == other {
		return 0
	}

	if q < other {
		return -1
	}

	return 1
}

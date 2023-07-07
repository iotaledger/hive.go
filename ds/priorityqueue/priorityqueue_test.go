package priorityqueue_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/ds/priorityqueue"
)

func TestPriorityQueue(t *testing.T) {
	queue := priorityqueue.New[int, priority]()
	queue.Push(2, 2)
	queue.Push(3, 3)
	peekedElement, exists := queue.Peek()
	assert.True(t, exists)
	assert.Equal(t, 2, peekedElement, "wrong element")
	poppedElement, exists := queue.Pop()
	assert.True(t, exists)
	assert.Equal(t, 2, poppedElement, "wrong element")
	peekedElement, exists = queue.Peek()
	assert.True(t, exists)
	assert.Equal(t, 3, peekedElement, "wrong element")
	queue.Push(1, 1)
	peekedElement, exists = queue.Peek()
	assert.True(t, exists)
	assert.Equal(t, 1, peekedElement, "wrong element")
}

type priority int

func (q priority) CompareTo(other priority) int {
	if q == other {
		return 0
	}

	if q < other {
		return -1
	}

	return 1
}

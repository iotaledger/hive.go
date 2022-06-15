package priorityqueue_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/generics/priorityqueue"
)

func TestPriorityQueue(t *testing.T) {
	queue := priorityqueue.New[queueElement]()
	queue.Push(2)
	queue.Push(3)
	assert.Equal(t, queueElement(2), queue.Peek(), "wrong element")
	assert.Equal(t, queueElement(2), queue.Pop(), "wrong element")
	assert.Equal(t, queueElement(3), queue.Peek(), "wrong element")
	queue.Push(1)
	assert.Equal(t, queueElement(1), queue.Peek(), "wrong element")
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

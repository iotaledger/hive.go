package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThreadSafeStack_Push(t *testing.T) {
	stack := newThreadSafeStack()

	assert.Equal(t, stack.Size(), 0, "stack should initially be empty")
	stack.Push(1)
	assert.Equal(t, stack.Size(), 1, "wrong stack size")
	stack.Push(2)
	assert.Equal(t, stack.Size(), 2, "wrong stack size")
	stack.Push(3)
	assert.Equal(t, stack.Size(), 3, "wrong stack size")
}

func TestThreadSafeStack_Pop(t *testing.T) {
	stack := newThreadSafeStack()

	assert.Equal(t, stack.Pop(), nil, "stack should return nil when its empty")
	stack.Push(1)
	stack.Push(2)
	assert.Equal(t, stack.Size(), 2, "wrong stack size")
	assert.Equal(t, 2, stack.Pop(), "wrong element popped from stack")
	assert.Equal(t, stack.Size(), 1, "wrong stack size")
	assert.Equal(t, 1, stack.Pop(), "wrong element popped from stack")
	assert.Equal(t, stack.Size(), 0, "wrong stack size")
	assert.Equal(t, stack.Pop(), nil, "stack should return nil when its empty")
}

func TestThreadSafeStack_Peek(t *testing.T) {
	stack := newThreadSafeStack()

	assert.Equal(t, stack.Peek(), nil, "stack should return nil when its empty")
	stack.Push(1)
	assert.Equal(t, stack.Size(), 1, "wrong stack size")
	assert.Equal(t, stack.Peek(), 1, "wrong element at top of stack")
	assert.Equal(t, stack.Size(), 1, "wrong stack size")
	stack.Push(2)
	assert.Equal(t, stack.Size(), 2, "wrong stack size")
	assert.Equal(t, stack.Peek(), 2, "wrong element at top of stack")
	assert.Equal(t, stack.Size(), 2, "wrong stack size")
}

func TestThreadSafeStack_Clear(t *testing.T) {
	stack := newThreadSafeStack()

	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	assert.Equal(t, stack.Size(), 3, "wrong stack size")
	stack.Clear()
	assert.Equal(t, stack.Size(), 0, "wrong stack size")
	assert.Equal(t, stack.Peek(), nil, "stack should return nil when its empty")
	assert.Equal(t, stack.Pop(), nil, "stack should return nil when its empty")
}

func TestThreadSafeStack_Size(t *testing.T) {
	stack := newThreadSafeStack()

	assert.Equal(t, stack.Size(), 0, "wrong stack size")
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	assert.Equal(t, stack.Size(), 3, "wrong stack size")
	for i := 0; i < 10000; i++ {
		stack.Push(i)
	}
	assert.Equal(t, stack.Size(), 10003, "wrong stack size")
}

func TestThreadSafeStack_IsEmpty(t *testing.T) {
	stack := newThreadSafeStack()

	assert.True(t, stack.IsEmpty(), "stack should be empty")
	stack.Push(1)
	assert.False(t, stack.IsEmpty(), "stack should not be empty")
	stack.Push(2)
	stack.Push(3)
	assert.False(t, stack.IsEmpty(), "stack should not be empty")
	stack.Clear()
	assert.True(t, stack.IsEmpty(), "stack should be empty")
}

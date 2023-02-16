package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThreadSafeStack_Push(t *testing.T) {
	stack := newThreadSafeStack[int]()

	assert.Equal(t, stack.Size(), 0, "stack should initially be empty")
	stack.Push(1)
	assert.Equal(t, stack.Size(), 1, "wrong stack size")
	stack.Push(2)
	assert.Equal(t, stack.Size(), 2, "wrong stack size")
	stack.Push(3)
	assert.Equal(t, stack.Size(), 3, "wrong stack size")
}

func TestThreadSafeStack_Pop(t *testing.T) {
	stack := newThreadSafeStack[int]()

	_, exists := stack.Pop()
	assert.False(t, exists, "stack should return false when its empty")
	stack.Push(1)
	stack.Push(2)
	assert.Equal(t, stack.Size(), 2, "wrong stack size")
	value, exists := stack.Pop()
	assert.True(t, exists, "stack should return true if its not empty")
	assert.Equal(t, 2, value, "wrong element popped from stack")
	assert.Equal(t, stack.Size(), 1, "wrong stack size")
	value, exists = stack.Pop()
	assert.True(t, exists, "stack should return true if its not empty")
	assert.Equal(t, 1, value, "wrong element popped from stack")
	assert.Equal(t, stack.Size(), 0, "wrong stack size")
	_, exists = stack.Pop()
	assert.False(t, exists, "stack should return false when its empty")
}

func TestThreadSafeStack_Peek(t *testing.T) {
	stack := newThreadSafeStack[int]()

	_, exists := stack.Peek()
	assert.False(t, exists, "stack should return false when its empty")
	stack.Push(1)
	assert.Equal(t, stack.Size(), 1, "wrong stack size")
	value, exists := stack.Peek()
	assert.True(t, exists, "stack should return true if its not empty")
	assert.Equal(t, value, 1, "wrong element at top of stack")
	assert.Equal(t, stack.Size(), 1, "wrong stack size")
	stack.Push(2)
	assert.Equal(t, stack.Size(), 2, "wrong stack size")
	value, exists = stack.Peek()
	assert.True(t, exists, "stack should return true if its not empty")
	assert.Equal(t, value, 2, "wrong element at top of stack")
	assert.Equal(t, stack.Size(), 2, "wrong stack size")
}

func TestThreadSafeStack_Clear(t *testing.T) {
	stack := newThreadSafeStack[int]()

	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	assert.Equal(t, stack.Size(), 3, "wrong stack size")
	stack.Clear()
	assert.Equal(t, stack.Size(), 0, "wrong stack size")
	_, exists := stack.Peek()
	assert.False(t, exists, "stack should return false when its empty")
	_, exists = stack.Pop()
	assert.False(t, exists, "stack should return false when its empty")
}

func TestThreadSafeStack_Size(t *testing.T) {
	stack := newThreadSafeStack[int]()

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
	stack := newThreadSafeStack[int]()

	assert.True(t, stack.IsEmpty(), "stack should be empty")
	stack.Push(1)
	assert.False(t, stack.IsEmpty(), "stack should not be empty")
	stack.Push(2)
	stack.Push(3)
	assert.False(t, stack.IsEmpty(), "stack should not be empty")
	stack.Clear()
	assert.True(t, stack.IsEmpty(), "stack should be empty")
}

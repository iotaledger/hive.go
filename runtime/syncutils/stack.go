package syncutils

import (
	"container/list"
	"sync"
)

type Stack[T any] struct {
	elements       *list.List
	mutex          sync.RWMutex
	elementAdded   *sync.Cond
	elementRemoved *sync.Cond
}

func NewStack[T any]() (newStack *Stack[T]) {
	newStack = new(Stack[T])
	newStack.elements = list.New()
	newStack.elementAdded = sync.NewCond(&newStack.mutex)
	newStack.elementRemoved = sync.NewCond(&newStack.mutex)

	return
}

func (b *Stack[T]) Push(task T) {
	b.mutex.Lock()
	b.elements.PushBack(task)
	b.mutex.Unlock()

	b.elementAdded.Broadcast()
}

func (b *Stack[T]) Pop() (element T, success bool) {
	defer func() {
		if success {
			b.elementRemoved.Broadcast()
		}
	}()

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if success = b.elements.Len() != 0; !success {
		return
	}

	//nolint:forcetypeassert // false positive, we know that the element is of type T
	return b.elements.Remove(b.elements.Front()).(T), true
}

func (b *Stack[T]) Size() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return b.elements.Len()
}

func (b *Stack[T]) PopOrWait(waitCondition func() bool) (element T, success bool) {
	defer func() {
		if success {
			b.elementRemoved.Broadcast()
		}
	}()

	b.mutex.Lock()
	defer b.mutex.Unlock()

	for b.elements.Len() == 0 {
		if success = waitCondition(); !success {
			return
		}

		b.elementAdded.Wait()
	}

	//nolint:forcetypeassert // false positive, we know that the element is of type T
	return b.elements.Remove(b.elements.Front()).(T), true
}

func (b *Stack[T]) WaitIsEmpty() {
	b.WaitSizeIsBelow(1)
}

func (b *Stack[T]) WaitSizeIsBelow(threshold int) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for b.elements.Len() >= threshold {
		b.elementRemoved.Wait()
	}
}

func (b *Stack[T]) WaitSizeIsAbove(threshold int) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for b.elements.Len() <= threshold {
		b.elementAdded.Wait()
	}
}

func (b *Stack[T]) SignalShutdown() {
	b.elementAdded.Broadcast()
}

package walker

import (
	"container/list"

	"github.com/iotaledger/hive.go/ds/orderedmap"
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/lo"
)

// region Walker /////////////////////////////////////////////////////////////////////////////////////////////////////////

// Walker implements a generic data structure that simplifies walks over collections or data structures.
type Walker[T comparable] struct {
	stack           *list.List
	pushedElements  *orderedmap.OrderedMap[T, types.Empty]
	walkStopped     bool
	revisitElements bool
}

// New is the constructor of the Walker. It accepts an optional boolean flag that controls whether the Walker will visit
// the same Element multiple times.
func New[T comparable](revisitElements ...bool) *Walker[T] {
	return &Walker[T]{
		stack:           list.New(),
		pushedElements:  orderedmap.New[T, types.Empty](),
		revisitElements: len(revisitElements) > 0 && revisitElements[0],
	}
}

// HasNext returns true if the Walker has another element that shall be visited.
func (w *Walker[T]) HasNext() bool {
	return w.stack.Len() > 0 && !w.walkStopped
}

// Pushed returns true if the passed element was Pushed to the Walker.
func (w *Walker[T]) Pushed(element T) bool {
	return w.pushedElements.Has(element)
}

// Next returns the next element of the walk.
func (w *Walker[T]) Next() (nextElement T) {
	currentEntry := w.stack.Front()
	w.stack.Remove(currentEntry)

	//nolint:forcetypeassert // false positive, we know that the element is of type T
	return currentEntry.Value.(T)
}

// Push adds a new element to the walk, which can consequently be retrieved by calling the Next method.
func (w *Walker[T]) Push(nextElement T) (walker *Walker[T]) {
	if lo.Return2(w.pushedElements.Set(nextElement, types.Void)) && !w.revisitElements {
		return w
	}

	w.stack.PushBack(nextElement)

	return w
}

// PushAll adds new elements to the walk, which can consequently be retrieved by calling the Next method.
func (w *Walker[T]) PushAll(nextElements ...T) (walker *Walker[T]) {
	for _, nextElement := range nextElements {
		w.Push(nextElement)
	}

	return w
}

// PushFront adds a new element to the front of the queue, which can consequently be retrieved by calling the Next method.
func (w *Walker[T]) PushFront(nextElements ...T) (walker *Walker[T]) {
	for _, nextElement := range nextElements {
		if lo.Return2(w.pushedElements.Set(nextElement, types.Void)) && !w.revisitElements {
			return w
		}

		w.stack.PushFront(nextElement)
	}

	return w
}

// StopWalk aborts the walk and forces HasNext to always return false.
func (w *Walker[T]) StopWalk() {
	w.walkStopped = true
}

// WalkStopped returns true if the Walk has been stopped.
func (w *Walker[T]) WalkStopped() bool {
	return w.walkStopped
}

// Reset removes all queued elements and reset walkStopped.
func (w *Walker[T]) Reset() {
	w.stack.Init()
	w.pushedElements.Clear()
	w.walkStopped = false
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

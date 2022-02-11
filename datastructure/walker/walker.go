package walker

import (
	"container/list"

	"github.com/iotaledger/hive.go/datastructure/set"
)

// region Walker /////////////////////////////////////////////////////////////////////////////////////////////////////////

// Walker implements a generic data structure that simplifies walks over collections or data structures.
type Walker struct {
	stack           *list.List
	pushedElements  set.Set
	walkStopped     bool
	revisitElements bool
}

// New is the constructor of the Walker. It accepts an optional boolean flag that controls whether the Walker will visit
// the same Element multiple times.
func New(revisitElements ...bool) (walker *Walker) {
	return &Walker{
		stack:           list.New(),
		pushedElements:  set.New(),
		revisitElements: len(revisitElements) > 0 && revisitElements[0],
	}
}

// HasNext returns true if the Walker has another element that shall be visited.
func (w *Walker) HasNext() bool {
	return w.stack.Len() > 0 && !w.walkStopped
}

// Pushed returns true if the passed element was Pushed to the Walker.
func (w *Walker) Pushed(element interface{}) bool {
	return w.pushedElements.Has(element)
}

// Next returns the next element of the walk.
func (w *Walker) Next() (nextElement interface{}) {
	currentEntry := w.stack.Front()
	w.stack.Remove(currentEntry)

	return currentEntry.Value
}

// Push adds a new element to the walk, which can consequently be retrieved by calling the Next method.
func (w *Walker) Push(nextElement interface{}) (walker *Walker) {
	if !w.pushedElements.Add(nextElement) && !w.revisitElements {
		return w
	}

	w.stack.PushBack(nextElement)

	return w
}

// StopWalk aborts the walk and forces HasNext to always return false.
func (w *Walker) StopWalk() {
	w.walkStopped = true
}

// WalkStopped returns true if the Walk has been stopped.
func (w *Walker) WalkStopped() bool {
	return w.walkStopped
}

// Reset removes all queued elements.
func (w *Walker) Reset() {
	w.stack.Init()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

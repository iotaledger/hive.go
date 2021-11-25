package walker

import (
	"container/list"

	"github.com/iotaledger/hive.go/v2/datastructure/set"
)

// region Walker /////////////////////////////////////////////////////////////////////////////////////////////////////////

// Walker implements a generic data structure that simplifies walks over collections or data structures.
type Walker struct {
	stack        *list.List
	seenElements set.Set
	walkStopped  bool
}

// New is the constructor of the Walker. It accepts an optional boolean flag that controls whether the Walker will visit
// the same Element multiple times.
func New(revisitElements ...bool) (walker *Walker) {
	walker = &Walker{
		stack: list.New(),
	}

	if len(revisitElements) == 0 || !revisitElements[0] {
		walker.seenElements = set.New()
	}

	return
}

// HasNext returns true if the Walker has another element that shall be visited.
func (w *Walker) HasNext() bool {
	return w.stack.Len() > 0 && !w.walkStopped
}

// Next returns the next element of the walk.
func (w *Walker) Next() (nextElement interface{}) {
	currentEntry := w.stack.Front()
	w.stack.Remove(currentEntry)

	return currentEntry.Value
}

// Push adds a new element to the walk, which can consequently be retrieved by calling the Next method.
func (w *Walker) Push(nextElement interface{}) {
	if w.seenElements != nil && !w.seenElements.Add(nextElement) {
		return
	}

	w.stack.PushBack(nextElement)
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

package walker

import (
	"github.com/iotaledger/hive.go/datastructure/walker"
)

// region Walker /////////////////////////////////////////////////////////////////////////////////////////////////////////

// Walker implements a generic data structure that simplifies walks over collections or data structures.
type Walker[T any] struct {
	*walker.Walker
}

// New is the constructor of the Walker. It accepts an optional boolean flag that controls whether the Walker will visit
// the same Element multiple times.
func New[T any](revisitElements ...bool) *Walker[T] {
	return &Walker[T]{Walker: walker.New(revisitElements...)}
}

// HasNext returns true if the Walker has another element that shall be visited.
func (w *Walker[T]) HasNext() bool {
	return w.Walker.HasNext()
}

// Next returns the next element of the walk.
func (w *Walker[T]) Next() (nextElement T) {
	return w.Walker.Next().(T)
}

// Push adds a new element to the walk, which can consequently be retrieved by calling the Next method.
func (w *Walker[T]) Push(nextElement T) (walker *Walker[T]) {
	w.Walker.Push(nextElement)
	return w
}

// PushAll adds new elements to the walk, which can consequently be retrieved by calling the Next method.
func (w *Walker[T]) PushAll(nextElements ...T) (walker *Walker[T]) {
	for _, nextElement := range nextElements {
		w.Walker.Push(nextElement)
	}

	return w
}

// StopWalk aborts the walk and forces HasNext to always return false.
func (w *Walker[T]) StopWalk() {
	w.Walker.StopWalk()
}

// WalkStopped returns true if the Walk has been stopped.
func (w *Walker[T]) WalkStopped() bool {
	return w.Walker.WalkStopped()
}

// Reset removes all queued elements.
func (w *Walker[T]) Reset() {
	w.Walker.Reset()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

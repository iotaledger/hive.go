package reactive

// WaitGroup is a reactive Event that waits for a set of elements to be done and that allows to inspect the pending
// elements.
type WaitGroup[T comparable] interface {
	// Event returns the event that is triggered when all elements are done.
	Event

	// Add adds the given elements to the wait group.
	Add(elements ...T)

	// Done marks the given elements as done and triggers the wait group if all elements are done.
	Done(elements ...T)

	// Wait waits until all elements are done.
	Wait()

	// PendingElements returns the currently pending elements.
	PendingElements() ReadableSet[T]

	// Debug subscribes to the PendingElements and logs the state of the WaitGroup to the console whenever it changes.
	Debug(optStringer ...func(T) string) (unsubscribe func())
}

// NewWaitGroup creates a new WaitGroup.
func NewWaitGroup[T comparable](elements ...T) WaitGroup[T] {
	return newWaitGroup(elements...)
}

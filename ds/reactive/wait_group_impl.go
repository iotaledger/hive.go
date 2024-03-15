package reactive

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/iotaledger/hive.go/ds"
	"github.com/iotaledger/hive.go/lo"
)

// waitGroup is the default implementation of the WaitGroup interface.
type waitGroup[T comparable] struct {
	// Event embeds the Event that is triggered when all elements are done.
	Event

	// pendingElements contains the currently pending elements.
	pendingElements Set[T]

	// pendingElementsCounter is the thread-safe counter that keeps track of the number of pending elements.
	pendingElementsCounter atomic.Int32
}

// newWaitGroup creates a new wait group.
func newWaitGroup[T comparable](elements ...T) *waitGroup[T] {
	w := &waitGroup[T]{
		Event:           NewEvent(),
		pendingElements: NewSet[T](),
	}

	w.Add(elements...)

	return w
}

// Add adds the given elements to the wait group.
func (w *waitGroup[T]) Add(elements ...T) {
	// first increase the counter so that the trigger is not executed before all elements are added
	w.pendingElementsCounter.Add(int32(len(elements)))

	// then add the elements (and correct the counter if the elements are already present)
	for _, element := range elements {
		if !w.pendingElements.Add(element) {
			w.pendingElementsCounter.Add(-1)
		}
	}
}

// Done marks the given elements as done and triggers the wait group if all elements are done.
func (w *waitGroup[T]) Done(elements ...T) {
	for _, element := range elements {
		if w.pendingElements.Delete(element) && w.pendingElementsCounter.Add(-1) == 0 {
			w.Trigger()
		}
	}
}

// Wait waits until all elements are done.
func (w *waitGroup[T]) Wait() {
	var wg sync.WaitGroup

	wg.Add(1)
	w.OnTrigger(wg.Done)
	wg.Wait()
}

// PendingElements returns the currently pending elements.
func (w *waitGroup[T]) PendingElements() ReadableSet[T] {
	return w.pendingElements
}

// Debug subscribes to the PendingElements and logs the state of the WaitGroup to the console whenever it changes.
func (w *waitGroup[T]) Debug(optStringer ...func(T) string) (unsubscribe func()) {
	return w.pendingElements.OnUpdate(func(_ ds.SetMutations[T]) {
		pendingElementsString := "DONE"
		if pendingElements := w.pendingElements.ToSlice(); len(pendingElements) != 0 {
			stringer := lo.First(optStringer, func(element T) string {
				return fmt.Sprint(element)
			})

			pendingElementsString = strings.Join(lo.Map(pendingElements, stringer), ", ")
		}

		fmt.Println("Waiting:", pendingElementsString)
	})
}

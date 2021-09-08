package events

import (
	"sync"
)

// region Queue ////////////////////////////////////////////////////////////////////////////////////////////////////////

// Queue represents an Event
type Queue struct {
	queuedElements      []*queueElement
	queuedElementsMutex sync.Mutex
}

// NewQueue returns an empty Queue.
func NewQueue() *Queue {
	return (&Queue{}).clear()
}

// Queue enqueues an Event to be triggered later (using the Trigger function).
func (q *Queue) Queue(event *Event, params ...interface{}) {
	q.queuedElementsMutex.Lock()
	defer q.queuedElementsMutex.Unlock()

	q.queuedElements = append(q.queuedElements, &queueElement{
		event:  event,
		params: params,
	})
}

// Trigger triggers all queued Events and empties the Queue.
func (q *Queue) Trigger() {
	q.queuedElementsMutex.Lock()
	defer q.queuedElementsMutex.Unlock()

	for _, queuedElement := range q.queuedElements {
		queuedElement.event.Trigger(queuedElement.params...)
	}
	q.clear()
}

// Clear removes all elements from the Queue.
func (q *Queue) Clear() {
	q.queuedElementsMutex.Lock()
	defer q.queuedElementsMutex.Unlock()

	q.clear()
}

// clear removes all elements from the Queue without locking it.
func (q *Queue) clear() *Queue {
	q.queuedElements = make([]*queueElement, 0)

	return q
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region queueElement /////////////////////////////////////////////////////////////////////////////////////////////////

// queueElement is a struct that holds the information about a triggered Event.
type queueElement struct {
	event  *Event
	params []interface{}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

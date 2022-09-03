package event

import (
	"sync"
)

// region LinkableCollectionEvent //////////////////////////////////////////////////////////////////////////////////////

type LinkableCollectionEvent[A any, B any, C ptrLinkableCollectionType[B, C]] struct {
	linkedEvent *LinkableCollectionEvent[A, B, C]
	linkClosure *Closure[A]

	*Event[A]
}

func NewLinkableCollectionEvent[A any, B any, C ptrLinkableCollectionType[B, C]](collection C, updateCollectionCallback func(target C)) (newEvent *LinkableCollectionEvent[A, B, C]) {
	collection.onLinkUpdated(updateCollectionCallback)

	return &LinkableCollectionEvent[A, B, C]{
		Event: New[A](),
	}
}

func (e *LinkableCollectionEvent[A, B, C]) Link(link *LinkableCollectionEvent[A, B, C]) {
	if e.linkClosure != nil {
		e.linkedEvent.Detach(e.linkClosure)
	} else {
		e.linkClosure = NewClosure(e.Trigger)
	}

	e.linkedEvent = link
	e.linkedEvent.Hook(e.linkClosure)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region LinkableCollection ///////////////////////////////////////////////////////////////////////////////////////////

type LinkableCollection[A any, B ptrLinkableCollectionType[A, B]] struct {
	linkUpdated *Event[B]
	sync.Once
}

func (l *LinkableCollection[A, B]) onLinkUpdated(callback func(linkTarget B)) {
	l.linkUpdatedEvent().Hook(NewClosure(callback))
}

func (l *LinkableCollection[A, B]) LinkTo(optLinkTargets ...B) {
	if len(optLinkTargets) == 0 {
		return
	}

	l.linkUpdatedEvent().Trigger(optLinkTargets[0])
}

func (l *LinkableCollection[A, B]) linkUpdatedEvent() (linkUpdatedEvent *Event[B]) {
	l.Do(func() {
		l.linkUpdated = New[B]()
	})

	return l.linkUpdated
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region LinkableCollectionConstructor ////////////////////////////////////////////////////////////////////////////////

func LinkableCollectionConstructor[A any, B ptrLinkableCollectionType[A, B]](init func(B)) (constructor func(...B) B) {
	return func(optLinkTargets ...B) (events B) {
		events = new(A)
		init(events)
		events.LinkTo(optLinkTargets...)

		return events
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region types and interfaces /////////////////////////////////////////////////////////////////////////////////////////

type ptrType[A any] interface {
	*A
}

type ptrLinkableCollectionType[A any, B ptrType[A]] interface {
	*A

	linkableCollectionType[B]
}

type linkableCollectionType[B any] interface {
	LinkTo(optionalLinkTargets ...B)
	onLinkUpdated(callback func(B))
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

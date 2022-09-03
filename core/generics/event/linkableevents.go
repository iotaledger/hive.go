package event

import (
	"sync"
)

// region LinkableCollectionEvent //////////////////////////////////////////////////////////////////////////////////////

// LinkableCollectionEvent represents a special kind of Event that is part of a LinkableCollection of events.
type LinkableCollectionEvent[A any, B any, C ptrLinkableCollectionType[B, C]] struct {
	linkedEvent *LinkableCollectionEvent[A, B, C]
	linkClosure *Closure[A]

	*Event[A]
}

// NewLinkableCollectionEvent creates a new LinkableCollectionEvent.
func NewLinkableCollectionEvent[A any, B any, C ptrLinkableCollectionType[B, C]](collection C, updateCollectionCallback func(target C)) (newEvent *LinkableCollectionEvent[A, B, C]) {
	collection.OnLinkUpdated(updateCollectionCallback)

	return &LinkableCollectionEvent[A, B, C]{
		Event: New[A](),
	}
}

// LinkTo links the LinkableCollectionEvent to the given LinkableCollectionEvent.
func (e *LinkableCollectionEvent[A, B, C]) LinkTo(link *LinkableCollectionEvent[A, B, C]) {
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

// LinkableCollection can be embedded into collections of linkable Events to make the entire collection linkable.
type LinkableCollection[A any, B ptrLinkableCollectionType[A, B]] struct {
	linkUpdated *Event[B]
	sync.Once
}

// LinkTo offers a way to link the collection to another collection of the same type.
func (l *LinkableCollection[A, B]) LinkTo(optLinkTargets ...B) {
	if len(optLinkTargets) == 0 {
		return
	}

	l.linkUpdatedEvent().Trigger(optLinkTargets[0])
}

// OnLinkUpdated registers a callback to be called when the link to the referenced LinkableCollection is set or updated.
func (l *LinkableCollection[A, B]) OnLinkUpdated(callback func(linkTarget B)) {
	l.linkUpdatedEvent().Hook(NewClosure(callback))
}

// linkUpdatedEvent returns the linkUpdated Event.
func (l *LinkableCollection[A, B]) linkUpdatedEvent() (linkUpdatedEvent *Event[B]) {
	l.Do(func() {
		l.linkUpdated = New[B]()
	})

	return l.linkUpdated
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region LinkableCollectionConstructor ////////////////////////////////////////////////////////////////////////////////

// LinkableCollectionConstructor contains a constructor-factory for collections that are linkable.
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

// ptrType is a helper type to create a pointer type.
type ptrType[A any] interface {
	*A
}

// ptsLinkableCollectionType is a helper type to create a pointer to a linkableCollectionType.
type ptrLinkableCollectionType[A any, B ptrType[A]] interface {
	*A

	linkableCollectionType[B]
}

// linkableCollectionType is the interface for a LinkableCollection.
type linkableCollectionType[B any] interface {
	LinkTo(optionalLinkTargets ...B)
	OnLinkUpdated(callback func(B))
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

package event

import (
	"sync"
)

// region Linkable /////////////////////////////////////////////////////////////////////////////////////////////////////

// Linkable represents a special kind of Event that is part of a LinkableCollection of events.
type Linkable[A any, B any, C ptrLinkableCollectionType[B, C]] struct {
	linkedEvent *Linkable[A, B, C]
	linkClosure *Closure[A]

	*Event[A]
}

// NewLinkable creates a new Linkable.
func NewLinkable[A any, B any, C ptrLinkableCollectionType[B, C]]() (newEvent *Linkable[A, B, C]) {
	return &Linkable[A, B, C]{
		Event: New[A](),
	}
}

// LinkTo links the Linkable to the given Linkable.
func (e *Linkable[A, B, C]) LinkTo(optLinkTargets ...*Linkable[A, B, C]) {
	if len(optLinkTargets) == 0 {
		return
	}

	if e.linkClosure != nil {
		e.linkedEvent.Detach(e.linkClosure)
	} else {
		e.linkClosure = NewClosure(e.Trigger)
	}

	e.linkedEvent = optLinkTargets[0]
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

// onLinkUpdated registers a callback to be called when the link to the referenced LinkableCollection is set or updated.
func (l *LinkableCollection[A, B]) onLinkUpdated(callback func(linkTarget B)) {
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

// region NewLinkableCollection ////////////////////////////////////////////////////////////////////////////////////////

// NewLinkableCollection is a generic constructor factory for LinkableCollection objects.
func NewLinkableCollection[A any, B ptrLinkableCollectionType[A, B]](init func(B) func(B)) (constructor func(...B) B) {
	return func(optLinkTargets ...B) (events B) {
		events = new(A)
		events.onLinkUpdated(init(events))
		events.LinkTo(optLinkTargets...)

		return events
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region types and interfaces /////////////////////////////////////////////////////////////////////////////////////////

// linkableType is the interface for linkable elements.
type linkableType[B any] interface {
	LinkTo(optLinkTargets ...B)
}

// linkableCollectionType is the interface for a LinkableCollection.
type linkableCollectionType[B any] interface {
	onLinkUpdated(callback func(B))

	linkableType[B]
}

// ptrType is a helper type to create a pointer type.
type ptrType[A any] interface {
	*A
}

// ptsLinkableCollectionType is a helper type to create a pointer to a linkableCollectionType.
type ptrLinkableCollectionType[A any, B ptrType[A]] interface {
	*A

	linkableCollectionType[B]
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

//nolint:golint,revive // golint throws false positives with generics here
package event

import (
	"reflect"
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
	if len(optLinkTargets) == 0 || e.linkedEvent == optLinkTargets[0] || e == optLinkTargets[0] {
		return
	}

	if e.linkClosure != nil {
		e.linkedEvent.Detach(e.linkClosure)
	}

	if e.linkedEvent = optLinkTargets[0]; e.linkedEvent == nil {
		e.linkClosure = nil

		return
	}

	if e.linkClosure == nil {
		e.linkClosure = NewClosure(e.Trigger)
	}

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

// region LinkableConstructor ////////////////////////////////////////////////////////////////////////////////////////////

// LinkableConstructor returns the linkable constructor for the given type.
func LinkableConstructor[A any, B ptrLinkableCollectionType[A, B]](newFunc func() B) (constructor func(...B) B) {
	return func(optLinkTargets ...B) (self B) {
		self = newFunc()

		selfValue := reflect.ValueOf(self).Elem()
		self.onLinkUpdated(func(linkTarget B) {
			if linkTarget == nil {
				linkTarget = new(A)
			}

			linkTargetValue := reflect.ValueOf(linkTarget).Elem()

			for i := 0; i < selfValue.NumField(); i++ {
				if sourceField := selfValue.Field(i); sourceField.Kind() == reflect.Ptr {
					if linkTo := sourceField.MethodByName("LinkTo"); linkTo.IsValid() {
						linkTo.Call([]reflect.Value{linkTargetValue.Field(i)})
					}
				}
			}
		})
		self.LinkTo(optLinkTargets...)

		return self
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

// ptrLinkableCollectionType is a helper type to create a pointer to a linkableCollectionType.
type ptrLinkableCollectionType[A any, B ptrType[A]] interface {
	*A

	linkableCollectionType[B]
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

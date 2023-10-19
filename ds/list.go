package ds

// region List /////////////////////////////////////////////////////////////////////////////////////////////////////////

// List represents an interface for a doubly linked list.
type List[T any] interface {
	// Init initializes or clears the List.
	Init() List[T]

	// Front returns the first element of the List or nil if it is empty.
	Front() ListElement[T]

	// Back returns the last element of the List or nil if it is empty.
	Back() ListElement[T]

	// PushFront inserts and returns a new element with the given value at the front of the List.
	PushFront(value T) ListElement[T]

	// PushBack inserts and returns a new element with the given value at the back of the List.
	PushBack(value T) ListElement[T]

	// Remove removes the given element from the List and returns its value.
	Remove(element ListElement[T]) (removedValue T)

	// InsertBefore inserts and returns a new element with the given value immediately before the given position.
	InsertBefore(value T, position ListElement[T]) ListElement[T]

	// InsertAfter inserts and returns a new element with the given value immediately after the given position.
	InsertAfter(value T, position ListElement[T]) ListElement[T]

	// MoveToFront moves the given element to the front of the List.
	MoveToFront(element ListElement[T])

	// MoveToBack moves the given element to the back of the List.
	MoveToBack(element ListElement[T])

	// MoveBefore moves the given element before the given position.
	MoveBefore(element, position ListElement[T])

	// MoveAfter moves the given element after the given position.
	MoveAfter(element, position ListElement[T])

	// PushBackList inserts the values of the other List at the back of this List.
	PushBackList(other List[T])

	// PushFrontList inserts the values of the other List at the front of this List.
	PushFrontList(other List[T])

	// ForEach executes the given callback for the value of each element in the List. The iteration is aborted if the
	// callback returns an error.
	ForEach(callback func(value T) error) error

	// ForEachReverse executes the given callback for the value of each element in the List in reverse order. The
	// iteration is aborted if the callback returns an error.
	ForEachReverse(callback func(value T) error) error

	// Range executes the given callback for the value of each element in the List.
	Range(callback func(value T))

	// RangeReverse executes the given callback for the value of each element in the List in reverse order.
	RangeReverse(callback func(value T))

	// Values returns a slice of all values in the List.
	Values() []T

	// Len returns the number of elements in the List.
	Len() int
}

// NewList creates a new List (the optional lockFree parameter can be set to true to create a List that is not
// thread-safe).
func NewList[T any](lockFree ...bool) List[T] {
	if len(lockFree) > 0 && lockFree[0] {
		return newList[T]()
	}

	return newThreadSafeList[T]()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ListElement //////////////////////////////////////////////////////////////////////////////////////////////////

// ListElement represents an interface for a doubly linked list element.
type ListElement[T any] interface {
	// Prev returns the previous ListElement or nil.
	Prev() ListElement[T]

	// Next returns the next ListElement or nil.
	Next() ListElement[T]

	// Value returns the value of the ListElement.
	Value() T
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

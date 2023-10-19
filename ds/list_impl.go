package ds

import (
	"sync"
	"sync/atomic"
)

// region list /////////////////////////////////////////////////////////////////////////////////////////////////////////

// list implements the non-thread-safe version of the List interface.
type list[T any] struct {
	// root is the sentinel list element that marks the beginning and end of the list.
	root listElement[T]

	// len is the current list length excluding the sentinel element.
	len int
}

// newList returns a new list instance.
func newList[T any]() *list[T] {
	l := new(list[T])
	l.Init()

	return l
}

// Init initializes or clears the List.
func (l *list[T]) Init() List[T] {
	l.root.next.Store(&l.root)
	l.root.prev.Store(&l.root)
	l.len = 0

	return l
}

// Front returns the first element of the List or nil if it is empty.
func (l *list[T]) Front() ListElement[T] {
	if l.len == 0 {
		return nil
	}

	return l.root.next.Load()
}

// Back returns the last element of the List or nil if it is empty.
func (l *list[T]) Back() ListElement[T] {
	if l.len == 0 {
		return nil
	}

	return l.root.prev.Load()
}

// PushFront inserts and returns a new element with the given value at the front of the List.
func (l *list[T]) PushFront(value T) ListElement[T] {
	l.lazyInit()

	return l.insertValue(value, &l.root)
}

// PushBack inserts and returns a new element with the given value at the back of the List.
func (l *list[T]) PushBack(value T) ListElement[T] {
	l.lazyInit()

	return l.insertValue(value, l.root.prev.Load())
}

// Remove removes the given element from the List and returns its value.
func (l *list[T]) Remove(e ListElement[T]) T {
	typedElement, ok := e.(*listElement[T])
	if !ok {
		panic("unsupported ListElement type")
	}

	if typedElement.list.Load() == l {
		// if e.list == l, l must have been initialized when e was inserted
		// in l or l == nil (e is a zero Element) and l.remove will crash
		l.remove(typedElement)
	}

	return *typedElement.value.Load()
}

// InsertBefore inserts and returns a new element with the given value immediately before the given position.
func (l *list[T]) InsertBefore(value T, position ListElement[T]) ListElement[T] {
	positionTyped, ok := position.(*listElement[T])
	if !ok {
		panic("unsupported ListElement type")
	}

	if positionTyped.list.Load() != l {
		return nil
	}

	return l.insertValue(value, positionTyped.prev.Load())
}

// InsertAfter inserts and returns a new element with the given value immediately after the given position.
func (l *list[T]) InsertAfter(value T, position ListElement[T]) ListElement[T] {
	positionTyped, ok := position.(*listElement[T])
	if !ok {
		panic("unsupported ListElement type")
	}

	if positionTyped.list.Load() != l {
		return nil
	}

	return l.insertValue(value, positionTyped)
}

// MoveToFront moves the given element to the front of the List.
func (l *list[T]) MoveToFront(element ListElement[T]) {
	typedElement, ok := element.(*listElement[T])
	if !ok {
		panic("unsupported ListElement type")
	}

	if typedElement.list.Load() != l || l.root.next.Load() == element {
		return
	}

	l.move(typedElement, &l.root)
}

// MoveToBack moves the given element to the back of the List.
func (l *list[T]) MoveToBack(element ListElement[T]) {
	typedElement, ok := element.(*listElement[T])
	if !ok {
		panic("unsupported ListElement type")
	}

	if typedElement.list.Load() != l || l.root.prev.Load() == element {
		return
	}

	l.move(typedElement, l.root.prev.Load())
}

// MoveBefore moves the given element before the given position.
func (l *list[T]) MoveBefore(element, position ListElement[T]) {
	typedElement, ok := element.(*listElement[T])
	if !ok {
		panic("unsupported ListElement type")
	}

	positionTyped, ok := element.(*listElement[T])
	if !ok {
		panic("unsupported ListElement type")
	}

	if typedElement.list.Load() != l || element == position || positionTyped.list.Load() != l {
		return
	}

	l.move(typedElement, positionTyped.prev.Load())
}

// MoveAfter moves the given element after the given position.
func (l *list[T]) MoveAfter(element, position ListElement[T]) {
	typedElement, ok := element.(*listElement[T])
	if !ok {
		panic("unsupported ListElement type")
	}

	positionTyped, ok := element.(*listElement[T])
	if !ok {
		panic("unsupported ListElement type")
	}

	if typedElement.list.Load() != l || element == position || positionTyped.list.Load() != l {
		return
	}

	l.move(typedElement, positionTyped)
}

// PushBackList inserts the values of the other List at the back of this List.
func (l *list[T]) PushBackList(other List[T]) {
	l.lazyInit()
	for i, e := other.Len(), other.Front(); i > 0; i, e = i-1, e.Next() {
		typedElement, ok := e.(*listElement[T])
		if !ok {
			panic("unsupported ListElement type")
		}

		l.insertValue(*typedElement.value.Load(), l.root.prev.Load())
	}
}

// PushFrontList inserts the values of the other List at the front of this List.
func (l *list[T]) PushFrontList(other List[T]) {
	l.lazyInit()
	for i, e := other.Len(), other.Back(); i > 0; i, e = i-1, e.Prev() {
		typedElement, ok := e.(*listElement[T])
		if !ok {
			panic("unsupported ListElement type")
		}

		l.insertValue(*typedElement.value.Load(), &l.root)
	}
}

// ForEach executes the given callback for the value of each element in the List. The iteration is aborted if the
// callback returns an error.
func (l *list[T]) ForEach(callback func(value T) error) error {
	for element := l.Front(); element != nil; element = element.Next() {
		if err := callback(element.Value()); err != nil {
			return err
		}
	}

	return nil
}

// ForEachReverse executes the given callback for the value of each element in the List in reverse order. The iteration
// is aborted if the callback returns an error.
func (l *list[T]) ForEachReverse(callback func(value T) error) error {
	for element := l.Back(); element != nil; element = element.Prev() {
		if err := callback(element.Value()); err != nil {
			return err
		}
	}

	return nil
}

// Range executes the given callback for the value of each element in the List.
func (l *list[T]) Range(callback func(value T)) {
	for element := l.Front(); element != nil; element = element.Next() {
		callback(element.Value())
	}
}

// RangeReverse executes the given callback for the value of each element in the List in reverse order.
func (l *list[T]) RangeReverse(callback func(value T)) {
	for element := l.Back(); element != nil; element = element.Prev() {
		callback(element.Value())
	}
}

// Values returns a slice of all values in the List.
func (l *list[T]) Values() []T {
	values := make([]T, 0)

	l.Range(func(value T) {
		values = append(values, value)
	})

	return values
}

// Len returns the number of elements in the List.
func (l *list[T]) Len() int { return l.len }

// lazyInit lazily initializes a zero List value.
func (l *list[T]) lazyInit() {
	if l.root.next.Load() == nil {
		l.Init()
	}
}

// insert inserts e after at, increments l.len, and returns e.
func (l *list[T]) insert(e, at *listElement[T]) *listElement[T] {
	e.prev.Store(at)
	e.next.Store(at.next.Load())
	e.prev.Load().next.Store(e)
	e.next.Load().prev.Store(e)
	e.list.Store(l)
	l.len++

	return e
}

// insertValue is a convenience wrapper for insert(&Element{Value: v}, at).
func (l *list[T]) insertValue(v T, at *listElement[T]) *listElement[T] {
	newElement := new(listElement[T])
	newElement.value.Store(&v)

	return l.insert(newElement, at)
}

// remove removes e from its list, decrements l.len.
func (l *list[T]) remove(e *listElement[T]) {
	e.prev.Load().next.Store(e.next.Load())
	e.next.Load().prev.Store(e.prev.Load())
	e.next.Store(nil) // avoid memory leaks
	e.prev.Store(nil) // avoid memory leaks
	e.list.Store(nil)
	l.len--
}

// move moves e to next to at.
func (l *list[T]) move(e, at *listElement[T]) {
	if e == at {
		return
	}
	e.prev.Load().next.Store(e.next.Load())
	e.next.Load().prev.Store(e.prev.Load())

	e.prev.Store(at)
	e.next.Store(at.next.Load())
	e.prev.Load().next.Store(e)
	e.next.Load().prev.Store(e)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region threadSafeList ///////////////////////////////////////////////////////////////////////////////////////////////

// threadSafeList implements the List interface in a thread-safe way.
type threadSafeList[T any] struct {
	// list is the underlying list implementation.
	*list[T]

	// mutex is used to synchronize access to the list.
	mutex sync.RWMutex
}

// newThreadSafeList creates a new threadSafeList instance.
func newThreadSafeList[T any]() *threadSafeList[T] {
	return &threadSafeList[T]{
		list: newList[T](),
	}
}

// Init initializes or clears the List.
func (t *threadSafeList[T]) Init() List[T] {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.list.Init()
}

// Front returns the first element of the List or nil if it is empty.
func (t *threadSafeList[T]) Front() ListElement[T] {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.list.Front()
}

// Back returns the last element of the List or nil if it is empty.
func (t *threadSafeList[T]) Back() ListElement[T] {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.list.Back()
}

// PushFront inserts and returns a new element with the given value at the front of the List.
func (t *threadSafeList[T]) PushFront(value T) ListElement[T] {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.list.PushFront(value)
}

// PushBack inserts and returns a new element with the given value at the back of the List.
func (t *threadSafeList[T]) PushBack(value T) ListElement[T] {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.list.PushBack(value)
}

// Remove removes the given element from the List and returns its value.
func (t *threadSafeList[T]) Remove(element ListElement[T]) (removedValue T) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.list.Remove(element)
}

// InsertBefore inserts and returns a new element with the given value immediately before the given position.
func (t *threadSafeList[T]) InsertBefore(value T, position ListElement[T]) ListElement[T] {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.list.InsertBefore(value, position)
}

// InsertAfter inserts and returns a new element with the given value immediately after the given position.
func (t *threadSafeList[T]) InsertAfter(value T, position ListElement[T]) ListElement[T] {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.list.InsertAfter(value, position)
}

// MoveToFront moves the given element to the front of the List.
func (t *threadSafeList[T]) MoveToFront(element ListElement[T]) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.list.MoveToFront(element)
}

// MoveToBack moves the given element to the back of the List.
func (t *threadSafeList[T]) MoveToBack(element ListElement[T]) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.list.MoveToBack(element)
}

// MoveBefore moves the given element before the given position.
func (t *threadSafeList[T]) MoveBefore(element, position ListElement[T]) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.list.MoveBefore(element, position)
}

// MoveAfter moves the given element after the given position.
func (t *threadSafeList[T]) MoveAfter(element, position ListElement[T]) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.list.MoveAfter(element, position)
}

// PushBackList inserts the values of the other List at the back of this List.
func (t *threadSafeList[T]) PushBackList(other List[T]) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.list.PushBackList(other)
}

// PushFrontList inserts the values of the other List at the front of this List.
func (t *threadSafeList[T]) PushFrontList(other List[T]) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.list.PushFrontList(other)
}

// ForEach executes the given callback for the value of each element in the List. The iteration is aborted if the
// callback returns an error.
func (t *threadSafeList[T]) ForEach(callback func(value T) error) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.list.ForEach(callback)
}

// ForEachReverse executes the given callback for the value of each element in the List in reverse order. The iteration
// is aborted if the callback returns an error.
func (t *threadSafeList[T]) ForEachReverse(callback func(value T) error) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.list.ForEachReverse(callback)
}

// Range executes the given callback for the value of each element in the List.
func (t *threadSafeList[T]) Range(callback func(value T)) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	t.list.Range(callback)
}

// RangeReverse executes the given callback for the value of each element in the List in reverse order.
func (t *threadSafeList[T]) RangeReverse(callback func(value T)) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	t.list.RangeReverse(callback)
}

// Values returns a slice of all values in the List.
func (t *threadSafeList[T]) Values() []T {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.list.Values()
}

// Len returns the number of elements in the List.
func (t *threadSafeList[T]) Len() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.list.Len()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region listElement //////////////////////////////////////////////////////////////////////////////////////////////////

// listElement implements the ListElement interface.
type listElement[T any] struct {
	// next and prev are the list elements before and after this list element.
	next, prev atomic.Pointer[listElement[T]]

	// list is the list this element belongs to.
	list atomic.Pointer[list[T]]

	// value is the value of this list element.
	value atomic.Pointer[T]
}

// Prev returns the previous ListElement or nil.
func (l *listElement[T]) Prev() ListElement[T] {
	if p, elementList := l.prev.Load(), l.list.Load(); elementList != nil && p != &elementList.root {
		return p
	}

	return nil
}

// Next returns the next ListElement or nil.
func (l *listElement[T]) Next() ListElement[T] {
	if nextElement, elementList := l.next.Load(), l.list.Load(); elementList != nil && nextElement != &elementList.root {
		return nextElement
	}

	return nil
}

// Value returns the value of the ListElement.
func (l *listElement[T]) Value() T {
	value := l.value.Load()
	if value == nil {
		var zeroValue T
		return zeroValue
	}

	return *value
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

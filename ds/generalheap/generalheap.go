package generalheap

// region Heap ////////////////////////////////////////////////////////////////////////////////////////////////////

// Heap defines a heap based on times.
type Heap[Key Comparable[Key], Value any] []*HeapElement[Key, Value]

// Len is the number of elements in the collection.
func (h Heap[K, V]) Len() int {
	return len(h)
}

// Less reports whether the element with index i should sort before the element with index j.
func (h Heap[K, V]) Less(i, j int) bool {
	return h[i].Key.CompareTo(h[j].Key) < 0
}

// Swap swaps the elements with indexes i and j.
func (h Heap[K, V]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index, h[j].index = i, j
}

// Push adds x as the last element to the heap.
func (h *Heap[K, V]) Push(x interface{}) {
	//nolint:forcetypeassert // false positive, we know that the element is of type *HeapElement[K, V]
	data := x.(*HeapElement[K, V])
	*h = append(*h, data)
	data.index = len(*h) - 1
}

// Pop removes and returns the last element of the heap.
func (h *Heap[K, V]) Pop() interface{} {
	n := len(*h)
	data := (*h)[n-1]
	(*h)[n-1] = nil // avoid memory leak
	*h = (*h)[:n-1]
	data.index = -1

	return data
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region HeapElement /////////////////////////////////////////////////////////////////////////////////////////////////

type Comparable[T any] interface {
	CompareTo(other T) int
}

type HeapElement[K Comparable[K], V any] struct {
	// Value represents the value of the queued element.
	Value V
	// Key represents the time of the element to be used as a key.
	Key   K
	index int
}

func (h HeapElement[K, V]) Index() int {
	return h.index
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

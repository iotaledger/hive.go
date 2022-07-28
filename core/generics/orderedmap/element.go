package orderedmap

// Element defines the model of each element of the orderedMap.
type Element[K comparable, V any] struct {
	key   K
	value V
	prev  *Element[K, V]
	next  *Element[K, V]
}

package lo

import (
	"github.com/iotaledger/hive.go/constraints"
)

// Cond is a conditional statement that returns the trueValue if the condition is true and the falseValue otherwise.
func Cond[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}

	return falseValue
}

// Map iterates over elements of collection, applies the mapper function to each element
// and returns an array of modified TargetType elements.
func Map[SourceType any, TargetType any](source []SourceType, mapper func(SourceType) TargetType) (target []TargetType) {
	target = make([]TargetType, len(source))
	for i, value := range source {
		target[i] = mapper(value)
	}

	return target
}

// Flatten takes a slice of slices and returns a flattened slice.
//
// Given a slice of slices `slices`, flatten iterates over each slice in `slices`,
// and appends its elements to a new slice `result`. The resulting slice is then
// returned. If `slices` is empty, flatten returns an empty slice.
//
// Example:
//
//	slices := [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
//	flattened := lo.Flatten(slices)
//	// flattened is [1 2 3 4 5 6 7 8 9]
//
// Note that the input slice `slices` is not modified by this function.
func Flatten[V any](slices [][]V) []V {
	var result []V

	for _, slice := range slices {
		result = append(result, slice...)
	}

	return result
}

// Equal checks if two slices are equal.
//
// Example:
//
//	lo.Equal([]int{1, 2, 3}, []int{1, 2, 3}) // returns true
func Equal[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for i, value := range a {
		if value != b[i] {
			return false
		}
	}

	return true
}

// Reduce reduces collection to a value which is the accumulated result of running each element in collection
// through accumulator, where each successive invocation is supplied the return value of the previous.
func Reduce[T any, R any](collection []T, accumulator func(R, T) R, initial R) R {
	for _, item := range collection {
		initial = accumulator(initial, item)
	}

	return initial
}

// Filter iterates over elements of collection, returning an array of all elements predicate returns truthy for.
func Filter[V any](collection []V, predicate func(V) bool) []V {
	var result []V

	for _, item := range collection {
		if predicate(item) {
			result = append(result, item)
		}
	}

	return result
}

// KeyBy transforms a slice or an array of structs to a map based on a pivot callback.
func KeyBy[K comparable, V any](collection []V, iteratee func(V) K) map[K]V {
	result := make(map[K]V, len(collection))

	for _, v := range collection {
		k := iteratee(v)
		result[k] = v
	}

	return result
}

// KeyOnlyBy transforms a slice or an array of structs to a map containing the keys based on a pivot callback.
func KeyOnlyBy[K comparable, V any](collection []V, iteratee func(V) K) map[K]struct{} {
	result := make(map[K]struct{}, len(collection))

	for _, v := range collection {
		k := iteratee(v)
		result[k] = struct{}{}
	}

	return result
}

// FilterByValue iterates over the map, returning a map of all elements predicate returns truthy for.
func FilterByValue[K comparable, V any](collection map[K]V, predicate func(V) bool) map[K]V {
	result := make(map[K]V)
	for key, value := range collection {
		if predicate(value) {
			result[key] = value
		}
	}

	return result
}

// Keys creates an array of the map keys.
func Keys[K comparable, V any](in map[K]V) []K {
	result := make([]K, 0, len(in))

	for k := range in {
		result = append(result, k)
	}

	return result
}

// Values creates an array of the map values.
func Values[K comparable, V any](in map[K]V) []V {
	result := make([]V, 0, len(in))

	for _, v := range in {
		result = append(result, v)
	}

	return result
}

// First returns the first element of the slice or an optionally provided default value if the collection is empty (if
// no default value is provided, the zero value of the collection's element type is returned).
func First[T any](slice []T, optDefaultValue ...T) (firstElement T) {
	if len(slice) == 0 {
		if len(optDefaultValue) == 0 {
			return
		}

		return optDefaultValue[0]
	}

	return slice[0]
}

// Any returns the first element of the map or an optionally provided default value if the collection is empty (if no
// default value is provided, the zero value of the collection's element type is returned).
func Any[K comparable, V any](collection map[K]V, optDefaultValue ...V) (anyValue V) {
	for _, v := range collection {
		return v
	}

	if len(optDefaultValue) == 0 {
		return
	}

	return optDefaultValue[0]
}

// ForEach iterates over elements of collection and invokes iteratee for each element.
func ForEach[T any](collection []T, iteratee func(T)) {
	for _, item := range collection {
		iteratee(item)
	}
}

// ReduceProperty reduces collection to a value which is the accumulated result of running each element in collection
// through property resolver, which extracts a value to be reduced from the type,
// and then accumulator, where each successive invocation is supplied the return value of the previous.
func ReduceProperty[A, B, C any](collection []A, propertyResolver func(A) B, accumulator func(C, B) C, initial C) C {
	for _, item := range collection {
		initial = accumulator(initial, propertyResolver(item))
	}

	return initial
}

// Bind generates a call wrapper with the second parameter's value fixed.
func Bind[FirstParamType, ParamType, ReturnType any](secondParam ParamType, callback func(FirstParamType, ParamType) ReturnType) func(FirstParamType) ReturnType {
	return func(firstParam FirstParamType) ReturnType {
		return callback(firstParam, secondParam)
	}
}

// PanicOnErr panics of the seconds parameter is an error and returns the first parameter otherwise.
func PanicOnErr[T any](result T, err error) T {
	if err != nil {
		panic(err)
	}

	return result
}

func DropCount[T any](result T, _ int, err error) (T, error) {
	return result, err
}

// Max returns the maximum value of the collection.
func Max[T constraints.Ordered](collection ...T) T {
	var maxElem T
	if len(collection) == 0 {
		return maxElem
	}

	maxElem = collection[0]

	return Reduce(collection, func(max, value T) T {
		if Comparator(value, max) > 0 {
			return value
		}

		return max
	}, maxElem)
}

// Min returns the minimum value of the collection.
func Min[T constraints.Ordered](collection ...T) T {
	var minElem T
	if len(collection) == 0 {
		return minElem
	}

	minElem = collection[0]

	return Reduce(collection, func(min, value T) T {
		if Comparator(value, min) < 0 {
			return value
		}

		return min
	}, minElem)
}

// Sum returns the sum of the collection.
func Sum[T constraints.Numeric](collection ...T) T {
	var sumElem T

	return Reduce(collection, func(sum, value T) T {
		return sum + value
	}, sumElem)
}

// MergeMaps updates the base map with values from the update map and returns the extended base map.
func MergeMaps[K comparable, V any](base, update map[K]V) map[K]V {
	for k, v := range update {
		base[k] = v
	}

	return base
}

// CopySlice copies the base slice into copied and returns the copy.
func CopySlice[T any](base []T) (copied []T) {
	copied = make([]T, len(base))
	copy(copied, base)

	return copied
}

// Return1 returns the first parameter out of a set of variadic arguments.
func Return1[A any](a A, _ ...any) A {
	return a
}

// Return2 returns the second parameter out of a set of variadic arguments.
func Return2[A any](_ any, a A, _ ...any) A {
	return a
}

// Return3 returns the third parameter out of a set of variadic arguments.
func Return3[A any](_, _ any, a A, _ ...any) A {
	return a
}

// Return4 returns the 4th parameter out of a set of variadic arguments..
func Return4[A any](_, _, _ any, a A, _ ...any) A {
	return a
}

// Return5 returns the 5th parameter out of a set of variadic arguments.
func Return5[A any](_, _, _, _ any, a A, _ ...any) A {
	return a
}

// Compare returns -1, 0 or 1 if the first parameter is smaller, equal or greater than the second argument.
func Compare[A constraints.Ordered](a, b A) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// Void returns a function that discards the return argument of the given function.
func Void[A, B any](f func(A) B) func(A) {
	return func(a A) { f(a) }
}

func CloneSlice[T any, C constraints.Cloneable[T]](slice []C) []T {
	cpy := make([]T, len(slice))
	for i, elem := range slice {
		cpy[i] = elem.Clone()
	}

	return cpy
}

func CloneMap[K comparable, V any, C constraints.Cloneable[V]](in map[K]C) map[K]V {
	cpy := make(map[K]V, len(in))
	for k, v := range in {
		cpy[k] = v.Clone()
	}

	return cpy
}

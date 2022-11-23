package onchangemap

import (
	"fmt"
	"sync"

	"github.com/iotaledger/hive.go/core/generics/constraints"
	"github.com/iotaledger/hive.go/core/generics/lo"
)

// Item represents an item in the OnChangeMap.
type Item[K comparable, C constraints.ComparableStringer[K]] interface {
	ID() C
	Clone() Item[K, C]
}

// OnChangeMap is a map that executes a callback if the map or an item is modified,
// in case callbackEnabled is true.
type OnChangeMap[K comparable, C constraints.ComparableStringer[K], I Item[K, C]] struct {
	mutex sync.RWMutex

	m               map[K]I
	callback        func([]I) error
	callbackEnabled bool
}

// NewOnChangeMap creates a new OnChangeMap.
func NewOnChangeMap[K comparable, C constraints.ComparableStringer[K], I Item[K, C]](callback func([]I) error) *OnChangeMap[K, C, I] {
	return &OnChangeMap[K, C, I]{
		m:               make(map[K]I),
		callback:        callback,
		callbackEnabled: false,
	}
}

// CallbackEnabled sets whether executing the callback on change is active or not.
func (r *OnChangeMap[K, C, I]) CallbackEnabled(enabled bool) {
	r.callbackEnabled = enabled
}

// executeCallbackWithoutLocking calls the callback if callbackEnabled is true.
func (r *OnChangeMap[K, C, I]) executeCallbackWithoutLocking() error {
	if !r.callbackEnabled {
		return nil
	}

	if r.callback == nil {
		return nil
	}

	if err := r.callback(lo.Values(r.m)); err != nil {
		return fmt.Errorf("failed to execute callback in OnChangeMap: %w", err)
	}

	return nil
}

// ExecuteCallback calls the callback if callbackEnabled is true.
func (r *OnChangeMap[K, C, I]) ExecuteCallback() error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.executeCallbackWithoutLocking()
}

// All returns a copy of all items.
func (r *OnChangeMap[K, C, I]) All() map[K]I {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	itemsCopy := make(map[K]I, len(r.m))
	for k := range r.m {
		itemsCopy[k] = r.m[k].Clone().(I)
	}

	return itemsCopy
}

// Get returns a copy of an item.
func (r *OnChangeMap[K, C, I]) Get(id C) (I, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if _, exists := r.m[id.Key()]; !exists {
		//nolint:gocritic
		return *new(I), fmt.Errorf("unable to get item: \"%s\" does not exist in map", id)
	}

	return r.m[id.Key()].Clone().(I), nil
}

// Add adds an item to the map.
func (r *OnChangeMap[K, C, I]) Add(item I) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.m[item.ID().Key()]; exists {
		return fmt.Errorf("unable to add item: \"%s\" already exists in map", item.ID())
	}

	r.m[item.ID().Key()] = item

	return r.executeCallbackWithoutLocking()
}

// Modify modifies an item in the map and returns a copy.
func (r *OnChangeMap[K, C, I]) Modify(id C, callback func(item I) bool) (I, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	item, exists := r.m[id.Key()]
	if !exists {
		//nolint:gocritic
		return *new(I), fmt.Errorf("unable to modify item: \"%s\" does not exist in map", id)
	}

	if !callback(item) {
		return item.Clone().(I), nil
	}

	return item.Clone().(I), r.executeCallbackWithoutLocking()
}

// Delete removes an item from the map.
func (r *OnChangeMap[K, C, I]) Delete(id C) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.m[id.Key()]; !exists {
		return fmt.Errorf("unable to remove item: \"%s\" does not exist in map", id)
	}

	delete(r.m, id.Key())

	return r.executeCallbackWithoutLocking()
}

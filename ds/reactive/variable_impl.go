package reactive

import (
	"log/slog"
	"sync"
	"unsafe"

	"github.com/iotaledger/hive.go/ds"
	"github.com/iotaledger/hive.go/lo"
)

// region variable /////////////////////////////////////////////////////////////////////////////////////////////////////

// variable is the default implementation of the Variable interface.
type variable[Type comparable] struct {
	*readableVariable[Type]

	// transformationFunc is the function that is used to transform the value before it is stored.
	transformationFunc func(currentValue Type, newValue Type) Type

	// updateOrderMutex is used to make sure that write operations are executed sequentially (all subscribers are
	// notified before the next write operation is executed).
	updateOrderMutex sync.Mutex
}

// newVariable creates a new variable with an optional transformation function that can be used to rewrite the value
// before it is stored.
func newVariable[Type comparable](transformationFunc ...func(currentValue Type, newValue Type) Type) *variable[Type] {
	return &variable[Type]{
		transformationFunc: lo.First(transformationFunc, func(_ Type, newValue Type) Type { return newValue }),
		readableVariable:   newReadableVariable[Type](),
	}
}

// Init is a convenience function that acts as a setter for the variable that can be chained with the constructor.
func (v *variable[Type]) Init(value Type) Variable[Type] {
	v.Set(value)

	return v
}

// Set sets the new value and triggers the registered callbacks if the value has changed.
func (v *variable[Type]) Set(newValue Type) (previousValue Type) {
	return v.Compute(func(Type) Type { return newValue })
}

// Compute computes the new value based on the current value and triggers the registered callbacks if the value changed.
func (v *variable[Type]) Compute(computeFunc func(currentValue Type) Type) (previousValue Type) {
	v.updateOrderMutex.Lock()
	defer v.updateOrderMutex.Unlock()

	newValue, previousValue, updateID, registeredCallbacks := v.updateValue(computeFunc)

	for _, registeredCallback := range registeredCallbacks {
		if registeredCallback.LockExecution(updateID) {
			registeredCallback.Invoke(previousValue, newValue)
			registeredCallback.UnlockExecution()
		}
	}

	return previousValue
}

// DefaultTo atomically sets the new value to the given default value if the current value is the zero value and
// triggers the registered callbacks if the value has changed. It returns the new value and a boolean flag that
// indicates if the value was updated.
func (v *variable[Type]) DefaultTo(defaultValue Type) (newValue Type, updated bool) {
	v.Compute(func(currentValue Type) Type {
		if updated = currentValue == *new(Type); updated {
			newValue = defaultValue
		} else {
			newValue = currentValue
		}

		return newValue
	})

	return newValue, updated
}

// InheritFrom inherits the value from the given ReadableVariable.
func (v *variable[Type]) InheritFrom(other ReadableVariable[Type]) (unsubscribe func()) {
	return other.OnUpdate(func(_, newValue Type) {
		v.Set(newValue)
	}, true)
}

// DeriveValueFrom is a utility function that allows to derive a value from a newly created DerivedVariable.
// It returns a teardown function that unsubscribes the DerivedVariable from its inputs.
func (v *variable[Type]) DeriveValueFrom(source DerivedVariable[Type]) (teardown func()) {
	// no need to unsubscribe variable from source (it will no longer change and get garbage collected after
	// unsubscribing from its inputs)
	_ = v.InheritFrom(source)

	return source.Unsubscribe
}

// ToggleValue sets the value to the given value and returns a function that resets the value to its zero value.
func (v *variable[Type]) ToggleValue(value Type) (reset func()) {
	v.Set(value)

	return func() {
		v.Set(*new(Type))
	}
}

// updateValue atomically prepares the trigger by setting the new value and returning the new value, the previous value,
// the triggerID and the callbacks to trigger.
func (v *variable[Type]) updateValue(newValueGenerator func(Type) Type) (newValue, previousValue Type, triggerID uniqueID, callbacksToTrigger []*callback[func(prevValue, newValue Type)]) {
	v.valueMutex.Lock()
	defer v.valueMutex.Unlock()

	if previousValue, newValue = v.value, v.transformationFunc(v.value, newValueGenerator(v.value)); newValue != previousValue {
		v.value = newValue
		triggerID = v.uniqueUpdateID.Next()
		callbacksToTrigger = v.registeredCallbacks.Values()
	}

	return newValue, previousValue, triggerID, callbacksToTrigger
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region readableVariable /////////////////////////////////////////////////////////////////////////////////////////////

// readableVariable is the default implementation of the ReadableVariable interface.
type readableVariable[Type comparable] struct {
	// value holds the current value.
	value Type

	// registeredCallbacks holds the callbacks that are triggered when the value changes.
	registeredCallbacks ds.List[*callback[func(prevValue, newValue Type)]]

	// uniqueUpdateID is used to derive a unique identifier for each update.
	uniqueUpdateID uniqueID

	// valueMutex is used to ensure that access to the value is synchronized.
	valueMutex sync.RWMutex
}

// newReadableVariable creates a new readableVariable instance with an optional initial value.
func newReadableVariable[Type comparable](initialValue ...Type) *readableVariable[Type] {
	return &readableVariable[Type]{
		value:               lo.First(initialValue),
		registeredCallbacks: ds.NewList[*callback[func(prevValue, newValue Type)]](),
	}
}

// Get returns the current value.
func (r *readableVariable[Type]) Get() Type {
	r.valueMutex.RLock()
	defer r.valueMutex.RUnlock()

	return r.value
}

// Read executes the given function with the current value while read locking the variable.
func (r *readableVariable[Type]) Read(readFunc func(currentValue Type)) {
	r.valueMutex.RLock()
	defer r.valueMutex.RUnlock()

	readFunc(r.value)
}

// WithValue is a utility function that allows to set up dynamic behavior based on the latest value of the
// ReadableVariable which is torn down once the value changes again (or the returned teardown function is called).
// It accepts an optional condition that has to be satisfied for the setup function to be called.
func (r *readableVariable[Type]) WithValue(setup func(value Type) (teardown func()), condition ...func(Type) bool) (teardown func()) {
	return r.OnUpdateWithContext(func(_, value Type, unsubscribeOnUpdate func(setup func() (teardown func()))) {
		if len(condition) == 0 || condition[0](value) {
			unsubscribeOnUpdate(func() func() { return setup(value) })
		}
	}, true)
}

// WithNonEmptyValue is a utility function that allows to set up dynamic behavior based on the latest (non-empty)
// value of the ReadableVariable which is torn down once the value changes again (or the returned teardown function
// is called).
func (r *readableVariable[Type]) WithNonEmptyValue(setup func(value Type) (teardown func())) (teardown func()) {
	return r.WithValue(setup, func(t Type) bool { return t != *new(Type) })
}

// OnUpdate registers the given callback that is triggered when the value changes.
func (r *readableVariable[Type]) OnUpdate(callback func(prevValue, newValue Type), triggerWithInitialZeroValue ...bool) (unsubscribe func()) {
	r.valueMutex.Lock()

	currentValue := r.value
	createdCallback := newCallback[func(prevValue, newValue Type)](callback)
	callbackElement := r.registeredCallbacks.PushBack(createdCallback)

	// grab the execution lock before we unlock the mutex, so the callback cannot be triggered by another
	// thread updating the value before we have called the callback with the initial value
	createdCallback.LockExecution(r.uniqueUpdateID)
	defer createdCallback.UnlockExecution()

	r.valueMutex.Unlock()

	var emptyValue Type
	if currentValue != emptyValue || lo.First(triggerWithInitialZeroValue) {
		createdCallback.Invoke(emptyValue, currentValue)
	}

	return func() {
		r.registeredCallbacks.Remove(callbackElement)

		createdCallback.MarkUnsubscribed()
	}
}

// OnUpdateOnce registers the given callback for the next update and then automatically unsubscribes it. It is possible
// to provide an optional condition that has to be satisfied for the callback to be triggered.
func (r *readableVariable[Type]) OnUpdateOnce(callback func(oldValue Type, newValue Type), optCondition ...func(oldValue Type, newValue Type) bool) (unsubscribe func()) {
	callbackTriggered := NewEvent()
	var triggeredPreValue, triggeredNewValue Type

	unsubscribe = r.OnUpdate(func(prevValue, newValue Type) {
		if callbackTriggered.Get() {
			return
		}

		if len(optCondition) != 0 && !optCondition[0](prevValue, newValue) {
			return
		}

		triggeredPreValue = prevValue
		triggeredNewValue = newValue

		callbackTriggered.Trigger()
	})

	callbackTriggered.OnTrigger(func() {
		go unsubscribe()

		callback(triggeredPreValue, triggeredNewValue)
	})

	return unsubscribe
}

// OnUpdateWithContext registers the given callback that is triggered when the value changes. In contrast to the
// normal OnUpdate method, this method provides the old and new value as well as a withinContext function that can
// be used to create subscriptions that are automatically unsubscribed when the callback is triggered again. It is
// possible to return nil from withinContext to indicate that nothing should happen when the value changes again.
func (r *readableVariable[Type]) OnUpdateWithContext(callback func(oldValue, newValue Type, withinContext func(subscribe func() (unsubscribe func()))), triggerWithInitialZeroValue ...bool) (unsubscribe func()) {
	var previousUnsubscribedEvent Event

	unsubscribeFromVariable := r.OnUpdate(func(oldValue, newValue Type) {
		if previousUnsubscribedEvent != nil {
			previousUnsubscribedEvent.Trigger()
		}

		unsubscribedEvent := NewEvent()
		withinContext := func(subscribe func() func()) {
			if !unsubscribedEvent.WasTriggered() {
				if teardownSubscription := subscribe(); teardownSubscription != nil {
					unsubscribedEvent.OnTrigger(teardownSubscription)
				}
			}
		}

		callback(oldValue, newValue, withinContext)

		previousUnsubscribedEvent = unsubscribedEvent
	}, triggerWithInitialZeroValue...)

	return func() {
		unsubscribeFromVariable()

		if previousUnsubscribedEvent != nil {
			previousUnsubscribedEvent.Trigger()
		}
	}
}

// LogUpdates configures the Variable to emit logs about updates with the given logger and log level. An optional
// stringer function can be provided to log the value in a custom format.
func (r *readableVariable[Type]) LogUpdates(logger VariableLogReceiver, logLevel slog.Level, variableName string, stringer ...func(Type) string) (unsubscribe func()) {
	logMessage := variableName

	return logger.OnLogLevelActive(logLevel, func() (shutdown func()) {
		return r.OnUpdate(func(_, newValue Type) {
			if isNil(newValue) {
				logger.LogAttrs(logMessage, logLevel, slog.String("set", "nil"))
			} else if len(stringer) != 0 {
				logger.LogAttrs(logMessage, logLevel, slog.String("set", stringer[0](newValue)))
			} else {
				logger.LogAttrs(logMessage, logLevel, slog.Any("set", newValue))
			}
		})
	})
}

// isNil returns true if the given value is nil.
func isNil(value any) bool {
	return value == nil || (*[2]uintptr)(unsafe.Pointer(&value))[1] == 0
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region derivedVariable //////////////////////////////////////////////////////////////////////////////////////////////

// derivedVariable implements the DerivedVariable interface.
type derivedVariable[ValueType comparable] struct {
	// Variable is the Variable that holds the derived value.
	Variable[ValueType]

	// unsubscribe is the function that is used to unsubscribe the derivedVariable from the inputs.
	unsubscribe func()

	// unsubscribeOnce is used to make sure that the unsubscribe function is only called once.
	unsubscribeOnce sync.Once
}

// newDerivedVariable creates a new derivedVariable instance.
func newDerivedVariable[ValueType comparable](subscribe func(DerivedVariable[ValueType]) func(), initialValue ...ValueType) *derivedVariable[ValueType] {
	d := &derivedVariable[ValueType]{
		Variable: NewVariable[ValueType]().Init(lo.First(initialValue)),
	}

	d.unsubscribe = subscribe(d)

	return d
}

// Unsubscribe unsubscribes the DerivedVariable from its input values.
func (d *derivedVariable[ValueType]) Unsubscribe() {
	d.unsubscribeOnce.Do(d.unsubscribe)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

package reactive

import (
	"log/slog"

	"github.com/iotaledger/hive.go/lo"
)

// region Variable /////////////////////////////////////////////////////////////////////////////////////////////////////

// Variable represents a variable that can be read and written and that informs subscribed consumers about updates.
type Variable[Type comparable] interface {
	// Init is a convenience function that acts as a setter for the variable that can be chained with the constructor.
	Init(value Type) Variable[Type]

	// WritableVariable imports the write methods of the Variable.
	WritableVariable[Type]

	// ReadableVariable imports the read methods of the Variable.
	ReadableVariable[Type]
}

// NewVariable creates a new Variable instance with an optional transformation function that can be used to rewrite the
// value before it is stored.
func NewVariable[Type comparable](transformationFunc ...func(currentValue Type, newValue Type) Type) Variable[Type] {
	return newVariable(transformationFunc...)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ReadableVariable /////////////////////////////////////////////////////////////////////////////////////////////

// ReadableVariable represents a variable that can be read and that informs subscribed consumers about updates.
type ReadableVariable[Type comparable] interface {
	// Get returns the current value.
	Get() Type

	// Read executes the given function with the current value while read locking the variable.
	Read(readFunc func(currentValue Type))

	// WithValue is a utility function that allows to set up dynamic behavior based on the latest value of the
	// ReadableVariable which is torn down once the value changes again (or the returned teardown function is called).
	// It accepts an optional condition that has to be satisfied for the setup function to be called.
	WithValue(setup func(value Type) (teardown func()), condition ...func(Type) bool) (teardown func())

	// WithNonEmptyValue is a utility function that allows to set up dynamic behavior based on the latest (non-empty)
	// value of the ReadableVariable which is torn down once the value changes again (or the returned teardown function
	// is called).
	WithNonEmptyValue(setup func(value Type) (teardown func())) (teardown func())

	// OnUpdate registers the given callback that is triggered when the value changes.
	OnUpdate(consumer func(oldValue, newValue Type), triggerWithInitialZeroValue ...bool) (unsubscribe func())

	// OnUpdateOnce registers the given callback for the next update and then automatically unsubscribes it. It is
	// possible to provide an optional condition that has to be satisfied for the callback to be triggered.
	OnUpdateOnce(callback func(oldValue, newValue Type), optCondition ...func(oldValue Type, newValue Type) bool) (unsubscribe func())

	// OnUpdateWithContext registers the given callback that is triggered when the value changes. In contrast to the
	// normal OnUpdate method, this method provides the old and new value as well as a withinContext function that can
	// be used to create subscriptions that are automatically unsubscribed when the callback is triggered again.
	OnUpdateWithContext(callback func(oldValue, newValue Type, withinContext func(subscriptionFactory func() (unsubscribe func()))), triggerWithInitialZeroValue ...bool) (unsubscribe func())

	// LogUpdates configures the Variable to emit logs about updates with the given logger and log level. An optional
	// stringer function can be provided to log the value in a custom format.
	LogUpdates(logger VariableLogReceiver, logLevel slog.Level, variableName string, stringer ...func(Type) string) (unsubscribe func())
}

// NewReadableVariable creates a new ReadableVariable instance with the given value.
func NewReadableVariable[Type comparable](value Type) ReadableVariable[Type] {
	return newReadableVariable(value)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region WritableVariable /////////////////////////////////////////////////////////////////////////////////////////////

// WritableVariable represents a variable that can be written to.
type WritableVariable[Type comparable] interface {
	// Set sets the new value and triggers the registered callbacks if the value has changed.
	Set(newValue Type) (previousValue Type)

	// Compute sets the new value by applying the given function to the current value and triggers the registered
	// callbacks if the value has changed.
	Compute(computeFunc func(currentValue Type) Type) (previousValue Type)

	// DefaultTo atomically sets the new value to the default value if the current value is the zero value and triggers
	// the registered callbacks if the value has changed. It returns the new value and a boolean flag that indicates if
	// the value was updated.
	DefaultTo(defaultValue Type) (newValue Type, updated bool)

	// InheritFrom inherits the value from the given ReadableVariable.
	InheritFrom(other ReadableVariable[Type]) (unsubscribe func())

	// DeriveValueFrom is a utility function that allows to derive a value from a newly created DerivedVariable.
	// It returns a teardown function that unsubscribes the DerivedVariable from its inputs.
	DeriveValueFrom(source DerivedVariable[Type]) (teardown func())

	// ToggleValue sets the value to the given value and returns a function that resets the value to its zero value.
	ToggleValue(value Type) (reset func())
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region DerivedVariable //////////////////////////////////////////////////////////////////////////////////////////////

// DerivedVariable is a Variable that automatically derives its value from other input values.
type DerivedVariable[Type comparable] interface {
	// Variable is the variable that holds the derived value.
	Variable[Type]

	// Unsubscribe unsubscribes the DerivedVariable from its input values.
	Unsubscribe()
}

// NewDerivedVariable creates a DerivedVariable that transforms an input value into a different one.
func NewDerivedVariable[Type, InputType1 comparable, InputValueType1 ReadableVariable[InputType1]](compute func(currentValue Type, inputValue1 InputType1) Type, input1 InputValueType1, initialValue ...Type) DerivedVariable[Type] {
	return newDerivedVariable[Type](func(d DerivedVariable[Type]) func() {
		return input1.OnUpdate(func(_, input1 InputType1) {
			d.Compute(func(currentValue Type) Type { return compute(currentValue, input1) })
		}, true)
	}, initialValue...)
}

// NewDerivedVariable2 creates a DerivedVariable that transforms two input values into a different one.
func NewDerivedVariable2[Type, InputType1, InputType2 comparable, InputValueType1 ReadableVariable[InputType1], InputValueType2 ReadableVariable[InputType2]](compute func(currentValue Type, inputValue1 InputType1, inputValue2 InputType2) Type, input1 InputValueType1, input2 InputValueType2, initialValue ...Type) DerivedVariable[Type] {
	return newDerivedVariable[Type](func(d DerivedVariable[Type]) func() {
		return lo.Batch(
			input1.OnUpdate(func(_, input1 InputType1) {
				d.Compute(func(currentValue Type) Type { return compute(currentValue, input1, input2.Get()) })
			}, true),

			input2.OnUpdate(func(_, input2 InputType2) {
				d.Compute(func(currentValue Type) Type { return compute(currentValue, input1.Get(), input2) })
			}, true),
		)
	}, initialValue...)
}

// NewDerivedVariable3 creates a DerivedVariable that transforms three input values into a different one.
func NewDerivedVariable3[Type, InputType1, InputType2, InputType3 comparable, InputValueType1 ReadableVariable[InputType1], InputValueType2 ReadableVariable[InputType2], InputValueType3 ReadableVariable[InputType3]](compute func(currentValue Type, inputValue1 InputType1, inputValue2 InputType2, inputValue3 InputType3) Type, input1 InputValueType1, input2 InputValueType2, input3 InputValueType3, initialValue ...Type) DerivedVariable[Type] {
	return newDerivedVariable[Type](func(d DerivedVariable[Type]) func() {
		return lo.Batch(
			input1.OnUpdate(func(_, input1 InputType1) {
				d.Compute(func(currentValue Type) Type { return compute(currentValue, input1, input2.Get(), input3.Get()) })
			}, true),

			input2.OnUpdate(func(_, input2 InputType2) {
				d.Compute(func(currentValue Type) Type { return compute(currentValue, input1.Get(), input2, input3.Get()) })
			}, true),

			input3.OnUpdate(func(_, input3 InputType3) {
				d.Compute(func(currentValue Type) Type { return compute(currentValue, input1.Get(), input2.Get(), input3) })
			}, true),
		)
	}, initialValue...)
}

// NewDerivedVariable4 creates a DerivedVariable that transforms four input values into a different one.
func NewDerivedVariable4[Type, InputType1, InputType2, InputType3, InputType4 comparable, InputValueType1 ReadableVariable[InputType1], InputValueType2 ReadableVariable[InputType2], InputValueType3 ReadableVariable[InputType3], InputValueType4 ReadableVariable[InputType4]](compute func(currentValue Type, inputValue1 InputType1, inputValue2 InputType2, inputValue3 InputType3, inputValue4 InputType4) Type, input1 InputValueType1, input2 InputValueType2, input3 InputValueType3, input4 InputValueType4, initialValue ...Type) DerivedVariable[Type] {
	return newDerivedVariable[Type](func(d DerivedVariable[Type]) func() {
		return lo.Batch(
			input1.OnUpdate(func(_, input1 InputType1) {
				d.Compute(func(currentValue Type) Type {
					return compute(currentValue, input1, input2.Get(), input3.Get(), input4.Get())
				})
			}, true),

			input2.OnUpdate(func(_, input2 InputType2) {
				d.Compute(func(currentValue Type) Type {
					return compute(currentValue, input1.Get(), input2, input3.Get(), input4.Get())
				})
			}, true),

			input3.OnUpdate(func(_, input3 InputType3) {
				d.Compute(func(currentValue Type) Type {
					return compute(currentValue, input1.Get(), input2.Get(), input3, input4.Get())
				})
			}, true),

			input4.OnUpdate(func(_, input4 InputType4) {
				d.Compute(func(currentValue Type) Type {
					return compute(currentValue, input1.Get(), input2.Get(), input3.Get(), input4)
				})
			}, true),
		)
	}, initialValue...)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region VariableLogReceiver //////////////////////////////////////////////////////////////////////////////////////////

// VariableLogReceiver defines the interface that is required to receive log messages from a Variable.
type VariableLogReceiver interface {
	// OnLogLevelActive registers a callback that is triggered when the given log level is activated. The shutdown
	// function that is returned by the callback is automatically called when the log level is deactivated.
	OnLogLevelActive(logLevel slog.Level, setup func() (shutdown func())) (unsubscribe func())

	// LogAttrs emits a log message with the given log level and attributes.
	LogAttrs(msg string, level slog.Level, args ...slog.Attr)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

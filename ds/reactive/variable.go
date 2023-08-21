package reactive

import (
	"github.com/iotaledger/hive.go/lo"
)

// region Variable /////////////////////////////////////////////////////////////////////////////////////////////////////

// Variable represents a variable that can be read and written and that informs subscribed consumers about updates.
type Variable[Type comparable] interface {
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

	// OnUpdate registers the given callback that is triggered when the value changes.
	OnUpdate(consumer func(oldValue, newValue Type), triggerWithInitialZeroValue ...bool) (unsubscribe func())
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

	// InheritFrom inherits the value from the given ReadableVariable.
	InheritFrom(other ReadableVariable[Type]) (unsubscribe func())
}

// region DerivedVariable //////////////////////////////////////////////////////////////////////////////////////////////

// DerivedVariable is a Variable that automatically derives its value from other input values.
type DerivedVariable[Type comparable] interface {
	// Variable is the variable that holds the derived value.
	Variable[Type]

	// Unsubscribe unsubscribes the DerivedVariable from its input values.
	Unsubscribe()
}

// NewDerivedVariable creates a DerivedVariable that transforms an input value into a different one.
func NewDerivedVariable[Type, InputType1 comparable, InputValueType1 ReadableVariable[InputType1]](compute func(InputType1) Type, input1 InputValueType1) DerivedVariable[Type] {
	return newDerivedVariable[Type](func(d DerivedVariable[Type]) func() {
		return input1.OnUpdate(func(_, input1 InputType1) {
			d.Compute(func(_ Type) Type { return compute(input1) })
		}, true)
	})
}

// NewDerivedVariable2 creates a DerivedVariable that transforms two input values into a different one.
func NewDerivedVariable2[Type, InputType1, InputType2 comparable, InputValueType1 ReadableVariable[InputType1], InputValueType2 ReadableVariable[InputType2]](compute func(InputType1, InputType2) Type, input1 InputValueType1, input2 InputValueType2) DerivedVariable[Type] {
	return newDerivedVariable[Type](func(d DerivedVariable[Type]) func() {
		return lo.Batch(
			input1.OnUpdate(func(_, input1 InputType1) {
				d.Compute(func(_ Type) Type { return compute(input1, input2.Get()) })
			}, true),

			input2.OnUpdate(func(_, input2 InputType2) {
				d.Compute(func(_ Type) Type { return compute(input1.Get(), input2) })
			}, true),
		)
	})
}

// NewDerivedVariable3 creates a DerivedVariable that transforms three input values into a different one.
func NewDerivedVariable3[Type, InputType1, InputType2, InputType3 comparable, InputValueType1 ReadableVariable[InputType1], InputValueType2 ReadableVariable[InputType2], InputValueType3 ReadableVariable[InputType3]](compute func(InputType1, InputType2, InputType3) Type, input1 InputValueType1, input2 InputValueType2, input3 InputValueType3) DerivedVariable[Type] {
	return newDerivedVariable[Type](func(d DerivedVariable[Type]) func() {
		return lo.Batch(
			input1.OnUpdate(func(_, input1 InputType1) {
				d.Compute(func(_ Type) Type { return compute(input1, input2.Get(), input3.Get()) })
			}, true),

			input2.OnUpdate(func(_, input2 InputType2) {
				d.Compute(func(_ Type) Type { return compute(input1.Get(), input2, input3.Get()) })
			}, true),

			input3.OnUpdate(func(_, input3 InputType3) {
				d.Compute(func(_ Type) Type { return compute(input1.Get(), input2.Get(), input3) })
			}, true),
		)
	})
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

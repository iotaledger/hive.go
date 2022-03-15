package dataflow

// region DataFlow /////////////////////////////////////////////////////////////////////////////////////////////////////

// DataFlow represents a chain of commands where the next command gets executed by the previous one passing on a common
// object holding the shared state. The recursive nature of the calls causes acquired resources (e.g. CachedObjects) to
// be held until the full data flow terminates.
//
// This allows us to pass through unwrapped objects in most of the business logic which relaxes the stress on the
// caching layer and makes the code less verbose.
type DataFlow[T any] struct {
	step                int
	max                 int
	commands            []ChainedCommand[T]
	successCallback     Callback[T]
	abortCallback       Callback[T]
	errorCallback       ErrorCallback[T]
	terminationCallback Callback[T]
}

// New creates a new DataFlow from the given ChainedCommands.
func New[T any](commands ...ChainedCommand[T]) (dataFlow *DataFlow[T]) {
	return &DataFlow[T]{
		commands: commands,
		max:      len(commands),
		step:     -1,
	}
}

// Run executes the DataFlow with the given parameter. It aborts execution and returns an error if any of the chained
// commands returns an error.
//
// Note: A DataFlow can only be run a single.
func (d *DataFlow[T]) Run(param T) (err error) {
	d.step++
	if d.step >= d.max {
		d.triggerSuccessCallback(param)

		return nil
	}

	if err = d.commands[d.step](param, d.Run); err != nil {
		d.triggerErrorCallback(err, param)
	}

	if d.step < d.max {
		d.triggerAbortCallback(param)
	}

	d.triggerTerminationCallback(param)

	return err
}

func (d *DataFlow[T]) WithSuccessCallback(callback Callback[T]) *DataFlow[T] {
	d.successCallback = callback

	return d
}

func (d *DataFlow[T]) WithAbortCallback(callback Callback[T]) *DataFlow[T] {
	d.abortCallback = callback

	return d
}

func (d *DataFlow[T]) WithErrorCallback(callback ErrorCallback[T]) *DataFlow[T] {
	d.errorCallback = callback

	return d
}

func (d *DataFlow[T]) WithTerminationCallback(callback Callback[T]) *DataFlow[T] {
	d.terminationCallback = callback

	return d
}

// ChainedCommand is a method that exposes the DataFlow as a ChainedCommand - use without calling it (without
// parentheses).
func (d *DataFlow[T]) ChainedCommand(param T, next Next[T]) error {
	return d.appendCommand(func(param T, done Next[T]) error {
		done(param)

		return next(param)
	}).Run(param)
}

func (d *DataFlow[T]) triggerSuccessCallback(param T) {
	if d.successCallback != nil {
		d.successCallback(param)
		d.successCallback = nil
	}
}

func (d *DataFlow[T]) triggerErrorCallback(err error, param T) {
	if d.errorCallback != nil {
		d.errorCallback(err, param)
		d.errorCallback = nil
	}
}

func (d *DataFlow[T]) triggerTerminationCallback(param T) {
	if d.terminationCallback != nil {
		d.terminationCallback(param)
		d.terminationCallback = nil
	}
}

func (d *DataFlow[T]) triggerAbortCallback(param T) {
	if d.abortCallback != nil {
		d.abortCallback(param)
		d.abortCallback = nil
	}
}

func (d *DataFlow[T]) appendCommand(command ChainedCommand[T]) (self *DataFlow[T]) {
	d.commands = append(d.commands, command)
	d.max++

	return d
}

// code contract (ensure that the RunWithCallback function is a ChainedCommand which enables composition of DataFlows).
var _ ChainedCommand[int] = new(DataFlow[int]).ChainedCommand

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ChainedCommand ///////////////////////////////////////////////////////////////////////////////////////////////

// ChainedCommand represents the interface for the callbacks used in a DataFlow.
type ChainedCommand[T any] func(param T, next Next[T]) error

type Next[T any] func(param T) error

type Callback[T any] func(param T)

type ErrorCallback[T any] func(err error, param T)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

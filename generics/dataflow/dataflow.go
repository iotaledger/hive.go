package dataflow

// region DataFlow /////////////////////////////////////////////////////////////////////////////////////////////////////

// DataFlow represents a chain of commands where the next command gets executed by the previous one passing on a common
// object holding the shared state. The recursive nature of the calls causes acquired resources (e.g. CachedObjects) to
// be held until the full data flow terminates.
//
// This allows us to pass through unwrapped objects in most of the business logic which relaxes the stress on the
// caching layer and makes the code less verbose.
type DataFlow[T any] struct {
	step     int
	max      int
	commands []ChainedCommand[T]
}

// New creates a new DataFlow from the given ChainedCommands.
func New[T any](steps ...ChainedCommand[T]) (dataFlow *DataFlow[T]) {
	return &DataFlow[T]{
		commands: steps,
		max:      len(steps),
		step:     -1,
	}
}

// Run executes the DataFlow with the given parameter. It aborts execution and returns an error if any of the chained
// commands returns an error.
//
// Note: A DataFlow can only be run a single.
func (d *DataFlow[T]) Run(param T) error {
	d.step++
	if d.step >= d.max {
		return nil
	}

	return d.commands[d.step](param, d.Run)
}

// RunWithCallback executes the DataFlow with the given parameter and executes a callback if the execution succeeds.
// It aborts execution and returns an error if any of the chained commands returns an error.
//
// RunWithCallback is a ChainedCommand itself and the method (without calling it) can be passed into other DataFlows to
// enable composition.
//
// Note: A DataFlow can only be run a single.
func (d *DataFlow[T]) RunWithCallback(param T, callback func(param T) error) error {
	d.commands = append(d.commands, func(param T, _ func(param T) error) error {
		return callback(param)
	})
	d.max++

	return d.Run(param)
}

// code contract (ensure that the RunWithCallback function is a ChainedCommand which enables composition of DataFlows).
var _ ChainedCommand[int] = new(DataFlow[int]).RunWithCallback

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ChainedCommand ///////////////////////////////////////////////////////////////////////////////////////////////

// ChainedCommand represents the interface for the callbacks used in a DataFlow.
type ChainedCommand[T any] func(param T, next func(param T) error) error

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

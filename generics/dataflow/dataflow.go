package dataflow

// region DataFlow /////////////////////////////////////////////////////////////////////////////////////////////////////

type Dataflow[T any] []ChainedCommand[T]

func New[T any](steps ...ChainedCommand[T]) (dataFlow Dataflow[T]) {
	return steps
}

func (d Dataflow[T]) Run(param T) error {
	return (&d).call(param)
}

func (d Dataflow[T]) ChainedCommand() ChainedCommand[T] {
	return func(param T, next func(param T) error) error {
		return (append(d, func(param T, _ func(param T) error) error {
			return next(param)
		})).Run(param)
	}
}

func (d *Dataflow[T]) call(param T) (err error) {
	if len(*d) == 0 {
		return nil
	}

	return d.nextCommand()(param, d.call)
}

func (d *Dataflow[T]) nextCommand() (nextCommand ChainedCommand[T]) {
	nextCommand = (*d)[0]
	*d = (*d)[1:]

	return nextCommand
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ChainedCommand ///////////////////////////////////////////////////////////////////////////////////////////////

type ChainedCommand[T any] func(param T, next func(param T) error) error

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

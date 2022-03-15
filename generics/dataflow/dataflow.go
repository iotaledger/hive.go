package dataflow

// DataFlow represents a set of steps that are executed in a sequence.
type DataFlow[T any] struct {
	steps           []Step[T]
	errorCallback   ErrorCallback[T]
	successCallback SuccessCallback[T]
}

func New[T any](steps ...Step[T]) *DataFlow[T] {
	return &DataFlow[T]{
		steps: steps,
	}
}

func (d *DataFlow[T]) OnSuccess(callback SuccessCallback[T]) (dataFlow *DataFlow[T]) {
	d.successCallback = callback

	return d
}

func (d *DataFlow[T]) OnError(callback ErrorCallback[T]) (dataFlow *DataFlow[T]) {
	d.errorCallback = callback

	return d
}

func (d *DataFlow[T]) Run(params T) (success bool) {
	for _, step := range d.steps {
		if err := step(params); err != nil {
			if d.errorCallback != nil {
				d.errorCallback(err, params)
			}
			return false
		}
	}

	if d.successCallback != nil {
		d.successCallback(params)
	}

	return true
}

type Step[T any] func(params T) (err error)

type ErrorCallback[T any] func(err error, params T)

type SuccessCallback[T any] func(params T)

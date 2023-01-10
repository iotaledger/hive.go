package event

import "github.com/iotaledger/hive.go/core/workerpool"

type handler[T any] struct {
	callback func(T)
	wp       *workerpool.UnboundedWorkerPool
}

func newHandler[T any](callback func(T), wp *workerpool.UnboundedWorkerPool) *handler[T] {
	return &handler[T]{
		callback: callback,
		wp:       wp,
	}
}

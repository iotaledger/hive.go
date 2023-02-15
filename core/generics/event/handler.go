package event

import "github.com/iotaledger/hive.go/runtime/workerpool"

type handler[T any] struct {
	callback func(T)
	wp       *workerpool.WorkerPool
}

func newHandler[T any](callback func(T), wp *workerpool.WorkerPool) *handler[T] {
	return &handler[T]{
		callback: callback,
		wp:       wp,
	}
}

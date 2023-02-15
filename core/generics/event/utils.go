package event

import "github.com/iotaledger/hive.go/runtime/workerpool"

func Attach[T any](event *Linkable[T], callback func(event T), triggerMaxCount ...uint64) (detachFunc func()) {
	closure := NewClosure(callback)
	event.Attach(closure, triggerMaxCount...)

	return func() {
		event.Detach(closure)
	}
}

func AttachWithWorkerPool[T any](event *Linkable[T], callback func(event T), wp *workerpool.UnboundedWorkerPool, triggerMaxCount ...uint64) (detachFunc func()) {
	closure := NewClosure(callback)
	event.AttachWithWorkerPool(closure, wp, triggerMaxCount...)

	return func() {
		event.Detach(closure)
	}
}

func Hook[T any](event *Linkable[T], callback func(event T), triggerMaxCount ...uint64) (detachFunc func()) {
	closure := NewClosure(callback)
	event.Hook(closure, triggerMaxCount...)

	return func() {
		event.Detach(closure)
	}
}

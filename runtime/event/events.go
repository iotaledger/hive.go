// Code generated by go generate; DO NOT EDIT.
package event

// Event is an event with no generic parameters.
type Event struct {
	*event[func()]
}

// New creates a new event with no generic parameters.
func New(opts ...Option) *Event {
	return &Event{
		event: newEvent[func()](opts...),
	}
}

// Trigger invokes the hooked callbacks.
func (e *Event) Trigger() {
	if e.currentTriggerExceedsMaxTriggerCount() {
		return
	}

	e.hooks.ForEach(func(_ uint64, hook *Hook[func()]) bool {
		if hook.currentTriggerExceedsMaxTriggerCount() {
			hook.Unhook()

			return true
		}

		if e.preTriggerFunc != nil {
			e.preTriggerFunc()
		}

		if hook.preTriggerFunc != nil {
			hook.preTriggerFunc()
		}

		if workerPool := hook.WorkerPool(); workerPool != nil {
			workerPool.Submit(func() { hook.trigger() })
		} else {
			hook.trigger()
		}

		return true
	})
}

// LinkTo links the event to the given target event (nil unlinks).
func (e *Event) LinkTo(target *Event) {
	e.linkTo(target, e.Trigger)
}

// Event1 is an event with 1 generic parameters.
type Event1[T1 any] struct {
	*event[func(T1)]
}

// New1 creates a new event with 1 generic parameters.
func New1[T1 any](opts ...Option) *Event1[T1] {
	return &Event1[T1]{
		event: newEvent[func(T1)](opts...),
	}
}

// Trigger invokes the hooked callbacks with the given parameters.
func (e *Event1[T1]) Trigger(arg1 T1) {
	if e.currentTriggerExceedsMaxTriggerCount() {
		return
	}

	e.hooks.ForEach(func(_ uint64, hook *Hook[func(T1)]) bool {
		if hook.currentTriggerExceedsMaxTriggerCount() {
			hook.Unhook()

			return true
		}

		if e.preTriggerFunc != nil {
			e.preTriggerFunc(arg1)
		}

		if hook.preTriggerFunc != nil {
			hook.preTriggerFunc(arg1)
		}

		if workerPool := hook.WorkerPool(); workerPool != nil {
			workerPool.Submit(func() { hook.trigger(arg1) })
		} else {
			hook.trigger(arg1)
		}

		return true
	})
}

// LinkTo links the event to the given target event (nil unlinks).
func (e *Event1[T1]) LinkTo(target *Event1[T1]) {
	e.linkTo(target, e.Trigger)
}

// Event2 is an event with 2 generic parameters.
type Event2[T1, T2 any] struct {
	*event[func(T1, T2)]
}

// New2 creates a new event with 2 generic parameters.
func New2[T1, T2 any](opts ...Option) *Event2[T1, T2] {
	return &Event2[T1, T2]{
		event: newEvent[func(T1, T2)](opts...),
	}
}

// Trigger invokes the hooked callbacks with the given parameters.
func (e *Event2[T1, T2]) Trigger(arg1 T1, arg2 T2) {
	if e.currentTriggerExceedsMaxTriggerCount() {
		return
	}

	e.hooks.ForEach(func(_ uint64, hook *Hook[func(T1, T2)]) bool {
		if hook.currentTriggerExceedsMaxTriggerCount() {
			hook.Unhook()

			return true
		}

		if e.preTriggerFunc != nil {
			e.preTriggerFunc(arg1, arg2)
		}

		if hook.preTriggerFunc != nil {
			hook.preTriggerFunc(arg1, arg2)
		}

		if workerPool := hook.WorkerPool(); workerPool != nil {
			workerPool.Submit(func() { hook.trigger(arg1, arg2) })
		} else {
			hook.trigger(arg1, arg2)
		}

		return true
	})
}

// LinkTo links the event to the given target event (nil unlinks).
func (e *Event2[T1, T2]) LinkTo(target *Event2[T1, T2]) {
	e.linkTo(target, e.Trigger)
}

// Event3 is an event with 3 generic parameters.
type Event3[T1, T2, T3 any] struct {
	*event[func(T1, T2, T3)]
}

// New3 creates a new event with 3 generic parameters.
func New3[T1, T2, T3 any](opts ...Option) *Event3[T1, T2, T3] {
	return &Event3[T1, T2, T3]{
		event: newEvent[func(T1, T2, T3)](opts...),
	}
}

// Trigger invokes the hooked callbacks with the given parameters.
func (e *Event3[T1, T2, T3]) Trigger(arg1 T1, arg2 T2, arg3 T3) {
	if e.currentTriggerExceedsMaxTriggerCount() {
		return
	}

	e.hooks.ForEach(func(_ uint64, hook *Hook[func(T1, T2, T3)]) bool {
		if hook.currentTriggerExceedsMaxTriggerCount() {
			hook.Unhook()

			return true
		}

		if e.preTriggerFunc != nil {
			e.preTriggerFunc(arg1, arg2, arg3)
		}

		if hook.preTriggerFunc != nil {
			hook.preTriggerFunc(arg1, arg2, arg3)
		}

		if workerPool := hook.WorkerPool(); workerPool != nil {
			workerPool.Submit(func() { hook.trigger(arg1, arg2, arg3) })
		} else {
			hook.trigger(arg1, arg2, arg3)
		}

		return true
	})
}

// LinkTo links the event to the given target event (nil unlinks).
func (e *Event3[T1, T2, T3]) LinkTo(target *Event3[T1, T2, T3]) {
	e.linkTo(target, e.Trigger)
}

// Event4 is an event with 4 generic parameters.
type Event4[T1, T2, T3, T4 any] struct {
	*event[func(T1, T2, T3, T4)]
}

// New4 creates a new event with 4 generic parameters.
func New4[T1, T2, T3, T4 any](opts ...Option) *Event4[T1, T2, T3, T4] {
	return &Event4[T1, T2, T3, T4]{
		event: newEvent[func(T1, T2, T3, T4)](opts...),
	}
}

// Trigger invokes the hooked callbacks with the given parameters.
func (e *Event4[T1, T2, T3, T4]) Trigger(arg1 T1, arg2 T2, arg3 T3, arg4 T4) {
	if e.currentTriggerExceedsMaxTriggerCount() {
		return
	}

	e.hooks.ForEach(func(_ uint64, hook *Hook[func(T1, T2, T3, T4)]) bool {
		if hook.currentTriggerExceedsMaxTriggerCount() {
			hook.Unhook()

			return true
		}

		if e.preTriggerFunc != nil {
			e.preTriggerFunc(arg1, arg2, arg3, arg4)
		}

		if hook.preTriggerFunc != nil {
			hook.preTriggerFunc(arg1, arg2, arg3, arg4)
		}

		if workerPool := hook.WorkerPool(); workerPool != nil {
			workerPool.Submit(func() { hook.trigger(arg1, arg2, arg3, arg4) })
		} else {
			hook.trigger(arg1, arg2, arg3, arg4)
		}

		return true
	})
}

// LinkTo links the event to the given target event (nil unlinks).
func (e *Event4[T1, T2, T3, T4]) LinkTo(target *Event4[T1, T2, T3, T4]) {
	e.linkTo(target, e.Trigger)
}

// Event5 is an event with 5 generic parameters.
type Event5[T1, T2, T3, T4, T5 any] struct {
	*event[func(T1, T2, T3, T4, T5)]
}

// New5 creates a new event with 5 generic parameters.
func New5[T1, T2, T3, T4, T5 any](opts ...Option) *Event5[T1, T2, T3, T4, T5] {
	return &Event5[T1, T2, T3, T4, T5]{
		event: newEvent[func(T1, T2, T3, T4, T5)](opts...),
	}
}

// Trigger invokes the hooked callbacks with the given parameters.
func (e *Event5[T1, T2, T3, T4, T5]) Trigger(arg1 T1, arg2 T2, arg3 T3, arg4 T4, arg5 T5) {
	if e.currentTriggerExceedsMaxTriggerCount() {
		return
	}

	e.hooks.ForEach(func(_ uint64, hook *Hook[func(T1, T2, T3, T4, T5)]) bool {
		if hook.currentTriggerExceedsMaxTriggerCount() {
			hook.Unhook()

			return true
		}

		if e.preTriggerFunc != nil {
			e.preTriggerFunc(arg1, arg2, arg3, arg4, arg5)
		}

		if hook.preTriggerFunc != nil {
			hook.preTriggerFunc(arg1, arg2, arg3, arg4, arg5)
		}

		if workerPool := hook.WorkerPool(); workerPool != nil {
			workerPool.Submit(func() { hook.trigger(arg1, arg2, arg3, arg4, arg5) })
		} else {
			hook.trigger(arg1, arg2, arg3, arg4, arg5)
		}

		return true
	})
}

// LinkTo links the event to the given target event (nil unlinks).
func (e *Event5[T1, T2, T3, T4, T5]) LinkTo(target *Event5[T1, T2, T3, T4, T5]) {
	e.linkTo(target, e.Trigger)
}

// Event6 is an event with 6 generic parameters.
type Event6[T1, T2, T3, T4, T5, T6 any] struct {
	*event[func(T1, T2, T3, T4, T5, T6)]
}

// New6 creates a new event with 6 generic parameters.
func New6[T1, T2, T3, T4, T5, T6 any](opts ...Option) *Event6[T1, T2, T3, T4, T5, T6] {
	return &Event6[T1, T2, T3, T4, T5, T6]{
		event: newEvent[func(T1, T2, T3, T4, T5, T6)](opts...),
	}
}

// Trigger invokes the hooked callbacks with the given parameters.
func (e *Event6[T1, T2, T3, T4, T5, T6]) Trigger(arg1 T1, arg2 T2, arg3 T3, arg4 T4, arg5 T5, arg6 T6) {
	if e.currentTriggerExceedsMaxTriggerCount() {
		return
	}

	e.hooks.ForEach(func(_ uint64, hook *Hook[func(T1, T2, T3, T4, T5, T6)]) bool {
		if hook.currentTriggerExceedsMaxTriggerCount() {
			hook.Unhook()

			return true
		}

		if e.preTriggerFunc != nil {
			e.preTriggerFunc(arg1, arg2, arg3, arg4, arg5, arg6)
		}

		if hook.preTriggerFunc != nil {
			hook.preTriggerFunc(arg1, arg2, arg3, arg4, arg5, arg6)
		}

		if workerPool := hook.WorkerPool(); workerPool != nil {
			workerPool.Submit(func() { hook.trigger(arg1, arg2, arg3, arg4, arg5, arg6) })
		} else {
			hook.trigger(arg1, arg2, arg3, arg4, arg5, arg6)
		}

		return true
	})
}

// LinkTo links the event to the given target event (nil unlinks).
func (e *Event6[T1, T2, T3, T4, T5, T6]) LinkTo(target *Event6[T1, T2, T3, T4, T5, T6]) {
	e.linkTo(target, e.Trigger)
}

// Event7 is an event with 7 generic parameters.
type Event7[T1, T2, T3, T4, T5, T6, T7 any] struct {
	*event[func(T1, T2, T3, T4, T5, T6, T7)]
}

// New7 creates a new event with 7 generic parameters.
func New7[T1, T2, T3, T4, T5, T6, T7 any](opts ...Option) *Event7[T1, T2, T3, T4, T5, T6, T7] {
	return &Event7[T1, T2, T3, T4, T5, T6, T7]{
		event: newEvent[func(T1, T2, T3, T4, T5, T6, T7)](opts...),
	}
}

// Trigger invokes the hooked callbacks with the given parameters.
func (e *Event7[T1, T2, T3, T4, T5, T6, T7]) Trigger(arg1 T1, arg2 T2, arg3 T3, arg4 T4, arg5 T5, arg6 T6, arg7 T7) {
	if e.currentTriggerExceedsMaxTriggerCount() {
		return
	}

	e.hooks.ForEach(func(_ uint64, hook *Hook[func(T1, T2, T3, T4, T5, T6, T7)]) bool {
		if hook.currentTriggerExceedsMaxTriggerCount() {
			hook.Unhook()

			return true
		}

		if e.preTriggerFunc != nil {
			e.preTriggerFunc(arg1, arg2, arg3, arg4, arg5, arg6, arg7)
		}

		if hook.preTriggerFunc != nil {
			hook.preTriggerFunc(arg1, arg2, arg3, arg4, arg5, arg6, arg7)
		}

		if workerPool := hook.WorkerPool(); workerPool != nil {
			workerPool.Submit(func() { hook.trigger(arg1, arg2, arg3, arg4, arg5, arg6, arg7) })
		} else {
			hook.trigger(arg1, arg2, arg3, arg4, arg5, arg6, arg7)
		}

		return true
	})
}

// LinkTo links the event to the given target event (nil unlinks).
func (e *Event7[T1, T2, T3, T4, T5, T6, T7]) LinkTo(target *Event7[T1, T2, T3, T4, T5, T6, T7]) {
	e.linkTo(target, e.Trigger)
}

// Event8 is an event with 8 generic parameters.
type Event8[T1, T2, T3, T4, T5, T6, T7, T8 any] struct {
	*event[func(T1, T2, T3, T4, T5, T6, T7, T8)]
}

// New8 creates a new event with 8 generic parameters.
func New8[T1, T2, T3, T4, T5, T6, T7, T8 any](opts ...Option) *Event8[T1, T2, T3, T4, T5, T6, T7, T8] {
	return &Event8[T1, T2, T3, T4, T5, T6, T7, T8]{
		event: newEvent[func(T1, T2, T3, T4, T5, T6, T7, T8)](opts...),
	}
}

// Trigger invokes the hooked callbacks with the given parameters.
func (e *Event8[T1, T2, T3, T4, T5, T6, T7, T8]) Trigger(arg1 T1, arg2 T2, arg3 T3, arg4 T4, arg5 T5, arg6 T6, arg7 T7, arg8 T8) {
	if e.currentTriggerExceedsMaxTriggerCount() {
		return
	}

	e.hooks.ForEach(func(_ uint64, hook *Hook[func(T1, T2, T3, T4, T5, T6, T7, T8)]) bool {
		if hook.currentTriggerExceedsMaxTriggerCount() {
			hook.Unhook()

			return true
		}

		if e.preTriggerFunc != nil {
			e.preTriggerFunc(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)
		}

		if hook.preTriggerFunc != nil {
			hook.preTriggerFunc(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)
		}

		if workerPool := hook.WorkerPool(); workerPool != nil {
			workerPool.Submit(func() { hook.trigger(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8) })
		} else {
			hook.trigger(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)
		}

		return true
	})
}

// LinkTo links the event to the given target event (nil unlinks).
func (e *Event8[T1, T2, T3, T4, T5, T6, T7, T8]) LinkTo(target *Event8[T1, T2, T3, T4, T5, T6, T7, T8]) {
	e.linkTo(target, e.Trigger)
}

// Event9 is an event with 9 generic parameters.
type Event9[T1, T2, T3, T4, T5, T6, T7, T8, T9 any] struct {
	*event[func(T1, T2, T3, T4, T5, T6, T7, T8, T9)]
}

// New9 creates a new event with 9 generic parameters.
func New9[T1, T2, T3, T4, T5, T6, T7, T8, T9 any](opts ...Option) *Event9[T1, T2, T3, T4, T5, T6, T7, T8, T9] {
	return &Event9[T1, T2, T3, T4, T5, T6, T7, T8, T9]{
		event: newEvent[func(T1, T2, T3, T4, T5, T6, T7, T8, T9)](opts...),
	}
}

// Trigger invokes the hooked callbacks with the given parameters.
func (e *Event9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) Trigger(arg1 T1, arg2 T2, arg3 T3, arg4 T4, arg5 T5, arg6 T6, arg7 T7, arg8 T8, arg9 T9) {
	if e.currentTriggerExceedsMaxTriggerCount() {
		return
	}

	e.hooks.ForEach(func(_ uint64, hook *Hook[func(T1, T2, T3, T4, T5, T6, T7, T8, T9)]) bool {
		if hook.currentTriggerExceedsMaxTriggerCount() {
			hook.Unhook()

			return true
		}

		if e.preTriggerFunc != nil {
			e.preTriggerFunc(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9)
		}

		if hook.preTriggerFunc != nil {
			hook.preTriggerFunc(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9)
		}

		if workerPool := hook.WorkerPool(); workerPool != nil {
			workerPool.Submit(func() { hook.trigger(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9) })
		} else {
			hook.trigger(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9)
		}

		return true
	})
}

// LinkTo links the event to the given target event (nil unlinks).
func (e *Event9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) LinkTo(target *Event9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) {
	e.linkTo(target, e.Trigger)
}

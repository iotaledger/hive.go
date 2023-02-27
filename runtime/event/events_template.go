//go:build ignore

package event

//go:generate go run github.com/iotaledger/hive.go/codegen/variadic/cmd@latest 0 9 events.go

// Event {{- if hasParams}}{{paramCount}}{{end}} is an event with {{if hasParams}}{{paramCount}}{{else}}no{{end}} generic parameters.
type Event /*{{- if hasParams}}{{paramCount}}{{"["}}{{types}}{{" any]"}}{{end}}*/ struct {
	*event[func( /*{{- types -}}*/ )]
}

// New {{- if hasParams}}{{paramCount}}{{end}} creates a new event with {{if hasParams}}{{paramCount}}{{else}}no{{end}} generic parameters.
func New /*{{- if hasParams}}{{paramCount}}{{"["}}{{types}}{{" any]"}}{{end -}}*/ (opts ...Option) *Event /*{{- if hasParams}}{{paramCount}}{{"["}}{{types}}{{"]"}}{{end}}*/ {
	return &Event /*{{- if hasParams}}{{paramCount}}{{"["}}{{types}}{{"]"}}{{end -}}*/ {
		event: newEvent[func( /*{{- types -}}*/ )](opts...),
	}
}

// Trigger invokes the hooked callbacks{{if hasParams}} with the given parameters{{end}}.
func (e *Event /*{{- if hasParams}}{{paramCount}}{{"["}}{{types}}{{"]"}}{{end}}*/) Trigger( /*{{- typedParams -}}*/ ) {
	if e.currentTriggerExceedsMaxTriggerCount() {
		return
	}

	e.hooks.ForEach(func(_ uint64, hook *Hook[func( /*{{- types -}}*/ )]) bool {
		if hook.currentTriggerExceedsMaxTriggerCount() {
			hook.Unhook()

			return true
		}

		if e.preTriggerFunc != nil {
			e.preTriggerFunc( /*{{- params -}}*/ )
		}

		if hook.preTriggerFunc != nil {
			hook.preTriggerFunc( /*{{- params -}}*/ )
		}

		if workerPool := hook.WorkerPool(); workerPool != nil {
			workerPool.Submit(func() { hook.trigger( /*{{- params -}}*/ ) })
		} else {
			hook.trigger( /*{{- params -}}*/ )
		}

		return true
	})
}

// LinkTo links the event to the given target event (nil unlinks).
func (e *Event /*{{- if hasParams}}{{paramCount}}{{"["}}{{types}}{{"]"}}{{end}}*/) LinkTo(target *Event /*{{- if hasParams}}{{paramCount}}{{"["}}{{types}}{{"]"}}{{end -}}*/) {
	e.linkTo(target, e.Trigger)
}

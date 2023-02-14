//go:build ignore

package event

//go:generate go run github.com/iotaledger/hive.go/codegen/variadic/cmd@e0630dc 0 9 events.go

// Event {{- if hasParams}}{{paramCount}}{{end}} is an event with {{if hasParams}}{{paramCount}}{{else}}no{{end}} generic parameters.
type Event /*{{- if hasParams}}{{paramCount}}{{"["}}{{types}}{{" any]"}}{{end}}*/ struct {
	*base[func( /*{{- types -}}*/ )]
}

// New {{- if hasParams}}{{paramCount}}{{end}} creates a new event with {{if hasParams}}{{paramCount}}{{else}}no{{end}} generic parameters.
func New /*{{- if hasParams}}{{paramCount}}{{"["}}{{types}}{{" any]"}}{{end -}}*/ (opts ...Option) *Event /*{{- if hasParams}}{{paramCount}}{{"["}}{{types}}{{"]"}}{{end}}*/ {
	return &Event /*{{- if hasParams}}{{paramCount}}{{"["}}{{types}}{{"]"}}{{end -}}*/ {
		base: newBase[func( /*{{- types -}}*/ )](opts...),
	}
}

// Trigger invokes the hooked callbacks{{if hasParams}} with the given parameters{{end}}.
func (e *Event /*{{- if hasParams}}{{paramCount}}{{"["}}{{types}}{{"]"}}{{end}}*/) Trigger( /*{{- typedParams -}}*/ ) {
	if e.MaxTriggerCountReached() {
		return
	}

	e.hooks.ForEach(func(_ uint64, hook *Hook[func( /*{{- types -}}*/ )]) bool {
		if hook.MaxTriggerCountReached() {
			hook.Unhook()

			return true
		}

		if workerPool := e.targetWorkerPool(hook); workerPool == nil {
			hook.trigger( /*{{- params -}}*/ )
		} else {
			workerPool.Submit(func() {
				hook.trigger( /*{{- params -}}*/ )
			})
		}

		return true
	})
}

// LinkTo links the event to the given target event (nil unlinks).
func (e *Event /*{{- if hasParams}}{{paramCount}}{{"["}}{{types}}{{"]"}}{{end}}*/) LinkTo(target *Event /*{{- if hasParams}}{{paramCount}}{{"["}}{{types}}{{"]"}}{{end -}}*/) {
	e.linkTo(target, e.Trigger)
}

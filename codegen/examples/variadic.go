package examples

import (
	"fmt"
)

//go:generate go run variadic_generate.go 5

// Event /*-paramCount*/ is an event with /*paramCount*/ generic parameters.
type Event /*-paramCount*/ /*-TypeParams*/ struct {
}

// New /*-paramCount*/ creates a new Event /*-paramCount*/ object.
func New /*-paramCount-*/ /*-TypeParams-*/ () *Event /*-paramCount*/ /*-typeParams*/ {
	return &Event /*-paramCount-*/ /*-typeParams*/ {}
}

// Trigger invokes the hooked callbacks with the given parameters.
func (e *Event /*-paramCount-*/ /*-typeParams-*/) Trigger( /*-Params-*/ ) {
	fmt.Println( /*-params-*/ )
}

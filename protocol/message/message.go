package message

import (
	"errors"
)

var (
	// ErrTypeAlreadyDefined is returned when an already defined message type is redefined.
	ErrTypeAlreadyDefined = errors.New("message type is already defined")
	// ErrUnknownType is returned when a definition for an unknown message type is requested.
	ErrUnknownType = errors.New("message type unknown")
)

// Type denotes the byte ID of a given message type.
type Type byte

// Definition describes a message's ID, its max byte length and whether its size can be variable.
type Definition struct {
	ID             Type
	MaxBytesLength uint16
	VariableLength bool
}

// Registry holds message definitions.
type Registry struct {
	definitions []*Definition
}

// NewRegistry create an empty message registry.
func NewRegistry() *Registry {
	return &Registry{definitions: make([]*Definition, 0)}
}

// Definitions returns all registered message definitions.
func (r *Registry) Definitions() []*Definition {
	return r.definitions
}

// RegisterType registers the given message type with its definition.
func (r *Registry) RegisterType(msgType Type, def *Definition) error {
	// grow definitions slice appropriately
	if len(r.definitions)-1 < int(msgType) {
		definitionsCopy := make([]*Definition, int(msgType)+1)
		copy(definitionsCopy, r.definitions)
		r.definitions = definitionsCopy
	}
	if r.definitions[msgType] != nil {
		return ErrTypeAlreadyDefined
	}
	r.definitions[msgType] = def
	return nil
}

// DefinitionForType returns the definition for the given message type.
func (r *Registry) DefinitionForType(msgType Type) (*Definition, error) {
	if len(r.definitions)-1 < int(msgType) {
		return nil, ErrUnknownType
	}
	def := r.definitions[msgType]
	if def == nil {
		return nil, ErrUnknownType
	}
	return def, nil
}

// Clear clears definitions of the registry
func (r *Registry) Clear() {
	r.definitions = make([]*Definition, 0)
}

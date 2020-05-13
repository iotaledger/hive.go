package message_test

import (
	"testing"

	"github.com/iotaledger/hive.go/protocol/message"
	"github.com/stretchr/testify/assert"
)

// message definition for testing
var (
	DummyMessageType         message.Type = 0
	DummyMessageDefinition = &message.Definition{
		ID:             DummyMessageType,
		MaxBytesLength: 10,
		VariableLength: false,
	}
)

func TestMessage_Register(t *testing.T) {
	r := message.NewRegistry()
	err := r.RegisterType(DummyMessageType, DummyMessageDefinition)
	assert.NoError(t, err)

	definitions := r.Definitions()
	assert.Equal(t, definitions[0], DummyMessageDefinition)
}

func TestMessage_RegisterTypeAlready(t *testing.T) {
	r := message.NewRegistry()
	err := r.RegisterType(DummyMessageType, DummyMessageDefinition)
	assert.NoError(t, err)
	err = r.RegisterType(DummyMessageType, DummyMessageDefinition)
	if assert.Error(t, err) {
		assert.Equal(t, message.ErrTypeAlreadyDefined, err)
	}
}

func TestMessage_DefinitionForType(t *testing.T) {
	r := message.NewRegistry()
	// registry is empty, len(definitions) = 0
	_, err := r.DefinitionForType(DummyMessageType)
	if assert.Error(t, err) {
		assert.Equal(t, message.ErrUnknownType, err)
	}

	// happy path
	err = r.RegisterType(DummyMessageType, DummyMessageDefinition)
	assert.NoError(t, err)

	var def *message.Definition
	def, err = r.DefinitionForType(DummyMessageType)
	assert.NoError(t, err)
	assert.Equal(t, DummyMessageDefinition, def)
}

func TestMessage_DefinitionForTypeBig(t *testing.T) {
	r := message.NewRegistry()
	// register a message with big type number
	var BigMessageType message.Type = 5
	var BigMessageDefinition = &message.Definition{
		ID:             BigMessageType,
		MaxBytesLength: 10,
		VariableLength: false,
	}
	// grows definitions to size of the message type (5)
	err := r.RegisterType(BigMessageType, BigMessageDefinition)
	assert.NoError(t, err)
	// Dummy was not registered
	_, err = r.DefinitionForType(DummyMessageType)
	if assert.Error(t, err) {
		assert.Equal(t, message.ErrUnknownType, err)
	}
}

func TestMessage_Clear(t *testing.T) {
	r := message.NewRegistry()
	err := r.RegisterType(DummyMessageType, DummyMessageDefinition)
	assert.NoError(t, err)

	definitions := r.Definitions()
	assert.Equal(t, definitions[0], DummyMessageDefinition)

	r.Clear()
	definitions = r.Definitions()
	assert.Equal(t, 0, len(definitions))
}

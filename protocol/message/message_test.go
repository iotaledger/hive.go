package message_test

import(
	"testing"

	"github.com/iotaledger/hive.go/protocol/message"
	"github.com/stretchr/testify/assert"
)

// message definition for testing
var DummyMessageType message.Type = 0
var DummyMessageDefinition = &message.Definition{
	ID:             DummyMessageType,
	MaxBytesLength: 10,
	VariableLength: false,
}

func TestMessage_Register(t *testing.T) {
	err := message.RegisterType(DummyMessageType, DummyMessageDefinition)
	assert.NoError(t, err)

	definitions := message.Definitions()
	assert.Equal(t, definitions[0], DummyMessageDefinition)
	message.ClearDefinitions()
}

func TestMessage_RegisterTypeAlready(t *testing.T) {
	err := message.RegisterType(DummyMessageType, DummyMessageDefinition)
	assert.NoError(t, err)
	err = message.RegisterType(DummyMessageType, DummyMessageDefinition)
	if assert.Error(t, err) {
		assert.Equal(t, message.ErrTypeAlreadyDefined, err)
	}
	message.ClearDefinitions()
}

func TestMessage_DefinitionForType(t *testing.T) {
	// registry is empty, len(definitions) = 0
	_, err := message.DefinitionForType(DummyMessageType)
	if assert.Error(t, err) {
		assert.Equal(t, message.ErrUnknownType, err)
	}

	// happy path
	err = message.RegisterType(DummyMessageType, DummyMessageDefinition)
	assert.NoError(t, err)

	var def *message.Definition
	def, err = message.DefinitionForType(DummyMessageType)
	assert.NoError(t, err)
	assert.Equal(t, DummyMessageDefinition, def)
	message.ClearDefinitions()

}

func TestMessage_DefinitionForTypeBig(t *testing.T) {
	// register a message with big type number
	var BigMessageType message.Type = 5
	var BigMessageDefinition = &message.Definition{
		ID:             BigMessageType,
		MaxBytesLength: 10,
		VariableLength: false,
	}
	// grows definitions to size of the message type (5)
	err := message.RegisterType(BigMessageType, BigMessageDefinition)
	assert.NoError(t, err)
	// Dummy was not registered
	_, err = message.DefinitionForType(DummyMessageType)
	if assert.Error(t, err) {
		assert.Equal(t, message.ErrUnknownType, err)
	}
	message.ClearDefinitions()
}
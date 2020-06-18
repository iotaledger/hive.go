package protocol

import (
	"errors"

	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/protocol/message"
	"github.com/iotaledger/hive.go/protocol/tlv"
)

// Events holds protocol related events.
type Events struct {
	// Holds event instances to attach to for received messages.
	// Use a message's ID to get the corresponding event.
	Received []*events.Event
	// Fired for generic protocol errors.
	Error *events.Event
}

// Protocol encapsulates the logic of parsing and sending protocol messages.
type Protocol struct {
	// Holds events for sent/received messages and generic errors.
	Events Events
	// message registry
	msgRegistry *message.Registry
	// the current receiving message
	receivingMessage *message.Definition
	// the buffer holding the receiving message data
	receiveBuffer []byte
	// the current offset within the receiving buffer
	receiveBufferOffset int
}

// New generates a new protocol instance which is ready to read a first message header.
func New(r *message.Registry) *Protocol {

	// load message definitions
	definitions := r.Definitions()

	// allocate event handlers for all message types
	receiveHandlers := make([]*events.Event, len(definitions))
	sentHandlers := make([]*events.Event, len(definitions))
	for i, def := range definitions {
		if def == nil {
			continue
		}
		receiveHandlers[i] = events.NewEvent(events.ByteSliceCaller)
		sentHandlers[i] = events.NewEvent(events.CallbackCaller)
	}

	protocol := &Protocol{
		msgRegistry: r,
		Events: Events{
			Received: receiveHandlers,
			Error:    events.NewEvent(events.ErrorCaller),
		},
		// the first message on the protocol is a TLV header
		receiveBuffer:    make([]byte, tlv.HeaderMessageDefinition.MaxBytesLength),
		receivingMessage: tlv.HeaderMessageDefinition,
	}

	return protocol
}

// Receive acts as an event handler for received data.
func (p *Protocol) Read(data []byte) (int, error) {
	offset := 0
	length := len(data)

	// continue to parse messages as long as we have data to consume
	for offset < length && p.receivingMessage != nil {

		// read in data into the receive buffer for the current message type
		bytesRead := byteutils.ReadAvailableBytesToBuffer(p.receiveBuffer, p.receiveBufferOffset, data, offset, length)

		p.receiveBufferOffset += bytesRead

		// advance consumed offset of received data
		offset += bytesRead

		if p.receiveBufferOffset != len(p.receiveBuffer) {
			return offset, nil
		}

		// message fully received
		p.receiveBufferOffset = 0

		// interpret the next message type if we received a header
		if p.receivingMessage.ID == tlv.HeaderMessageDefinition.ID {

			header, err := tlv.ParseHeader(p.receiveBuffer, p.msgRegistry)
			if err != nil {
				p.Events.Error.Trigger(err)
				return offset, errors.New("")
			}

			// advance to handle the message type the header says we are receiving
			p.receivingMessage = header.Definition

			// allocate enough space for it
			p.receiveBuffer = make([]byte, header.MessageBytesLength)
			continue
		}

		// fire the message type's event handler.
		// note that the message id is valid here because we verified that the message type
		// exists while parsing the TLV header
		p.Events.Received[p.receivingMessage.ID].Trigger(p.receiveBuffer)

		// reset to receiving a header
		p.receivingMessage = tlv.HeaderMessageDefinition
		p.receiveBuffer = make([]byte, tlv.HeaderMessageDefinition.MaxBytesLength)
	}

	return offset, nil
}

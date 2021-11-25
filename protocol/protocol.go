package protocol

import (
	"sync"

	"github.com/iotaledger/hive.go/v2/byteutils"
	"github.com/iotaledger/hive.go/v2/events"
	"github.com/iotaledger/hive.go/v2/protocol/message"
	"github.com/iotaledger/hive.go/v2/protocol/tlv"
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
	// lock during concurrent reads
	readMutex sync.Mutex
	// the current receiving message
	readMessage *message.Definition
	// the buffer holding the receiving message data
	readBuffer []byte
	// the current offset within the receiving buffer
	readBufferOffset int
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
		sentHandlers[i] = events.NewEvent(events.VoidCaller)
	}

	protocol := &Protocol{
		msgRegistry: r,
		Events: Events{
			Received: receiveHandlers,
			Error:    events.NewEvent(events.ErrorCaller),
		},
		// the first message on the protocol is a TLV header
		readBuffer:  make([]byte, tlv.HeaderMessageDefinition.MaxBytesLength),
		readMessage: tlv.HeaderMessageDefinition,
	}

	return protocol
}

// Read acts as an event handler for received data.
func (p *Protocol) Read(data []byte) (int, error) {
	p.readMutex.Lock()
	defer p.readMutex.Unlock()

	offset := 0
	length := len(data)

	// continue to parse messages as long as we have data to consume
	for offset < length && p.readMessage != nil {

		// read in data into the receive buffer for the current message type
		bytesRead := byteutils.ReadAvailableBytesToBuffer(p.readBuffer, p.readBufferOffset, data, offset, length)

		p.readBufferOffset += bytesRead

		// advance consumed offset of received data
		offset += bytesRead

		// we din't receive the full message yet
		if p.readBufferOffset != len(p.readBuffer) {
			return offset, nil
		}

		// message fully received
		p.readBufferOffset = 0

		// interpret the next message type if we received a header
		if p.readMessage.ID == tlv.HeaderMessageDefinition.ID {

			header, err := tlv.ParseHeader(p.readBuffer, p.msgRegistry)
			if err != nil {
				p.Events.Error.Trigger(err)
				return offset, err
			}

			// advance to handle the message type the header says we are receiving
			p.readMessage = header.Definition

			// allocate enough space for it
			p.readBuffer = make([]byte, header.MessageBytesLength)
			continue
		}

		// fire the message type's event handler.
		// note that the message id is valid here because we verified that the message type
		// exists while parsing the TLV header
		p.Events.Received[p.readMessage.ID].Trigger(p.readBuffer)

		// reset to receiving a header
		p.readMessage = tlv.HeaderMessageDefinition
		p.readBuffer = make([]byte, tlv.HeaderMessageDefinition.MaxBytesLength)
	}

	return offset, nil
}

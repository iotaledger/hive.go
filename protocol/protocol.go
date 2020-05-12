package protocol

import (
	"fmt"
	"io"

	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/syncutils"
	"github.com/iotaledger/hive.go/protocol/message"
	"github.com/iotaledger/hive.go/protocol/tlv"
)


// Events holds protocol related events.
type Events struct {
	// Holds event instances to attach to for received messages.
	// Use a message's ID to get the corresponding event.
	Received []*events.Event
	// Holds event instances to attach to for sent messages.
	// Use a message's ID to get the corresponding event.
	Sent []*events.Event
	// Fired for generic protocol errors.
	// It is suggested to close the underlying ReadWriteCloser of the Protocol instance
	// if any error occurs.
	Error *events.Event
}

// Protocol encapsulates the logic of parsing and sending protocol messages.
type Protocol struct {
	// Holds events for sent and received messages, handshake completion and generic errors.
	Events Events
	// the underlying connection
	conn io.ReadWriteCloser
	// the current receiving message
	receivingMessage *message.Definition
	// the buffer holding the receiving message data
	receiveBuffer []byte
	// the current offset within the receiving buffer
	receiveBufferOffset int
	// mutex to synchronize multiple sends
	sendMutex syncutils.Mutex
}

// New generates a new protocol instance which is ready to read a first message header.
func New(conn io.ReadWriteCloser) *Protocol {

	// load message definitions
	definitions := message.Definitions()

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
		conn: conn,
		Events: Events{
			Received:           receiveHandlers,
			Sent:               sentHandlers,
			Error:              events.NewEvent(events.ErrorCaller),
		},
		// the first message on the protocol is a TLV header
		receiveBuffer:    make([]byte, tlv.HeaderMessageDefinition.MaxBytesLength),
		receivingMessage: tlv.HeaderMessageDefinition,
	}

	return protocol
}

// Start kicks off the protocol by starting to read from the connection.
func (p *Protocol) Start() {
	// start reading from the connection
	_, _ = p.conn.Read(make([]byte, 2048))
}

// Receive acts as an event handler for received data.
func (p *Protocol) Receive(data []byte) {
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
			return
		}

		// message fully received
		p.receiveBufferOffset = 0

		// interpret the next message type if we received a header
		if p.receivingMessage.ID == tlv.HeaderMessageDefinition.ID {

			header, err := tlv.ParseHeader(p.receiveBuffer)
			if err != nil {
				p.Events.Error.Trigger(err)
				_ = p.conn.Close()
				return
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
}

// Send sends the given message (including the message header) to the underlying writer.
// It fires the corresponding send event for the specific message type.
func (p *Protocol) Send(message []byte) error {
	p.sendMutex.Lock()
	defer p.sendMutex.Unlock()

	// write message
	if _, err := p.conn.Write(message); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// fire event handler for sent message
	p.Events.Sent[message[0]].Trigger()

	return nil
}

// CloseConnection closes the underlying connection
func (p *Protocol) CloseConnection() {
	p.conn.Close()
}

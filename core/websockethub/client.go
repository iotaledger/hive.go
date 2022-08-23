package websockethub

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/atomic"
)

const (
	// time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// The message types are defined in RFC 6455, section 11.8.
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

// WebsocketMsg is a message received via websocket.
type WebsocketMsg struct {
	// MsgType is the type of the message based on RFC 6455.
	MsgType int
	// Data is the received data of the message.
	Data []byte
}

// Client is a middleman between the node and the websocket connection.
type Client struct {
	hub *Hub

	// the websocket connection.
	conn *websocket.Conn

	// a channel which is closed when the websocket client is disconnected.
	ExitSignal chan struct{}

	// buffered channel of outbound messages.
	sendChan chan interface{}

	// a channel which is closed when the writePump of the client exited.
	// this is used signal the hub to not send messages to sendChan anymore.
	sendChanClosed chan struct{}

	// channel of inbound messages.
	// this will be created by the user if receiving messages is needed.
	ReceiveChan chan *WebsocketMsg

	// onConnect gets called when the client was registered
	onConnect func(*Client)

	// FilterCallback is used to filter messages to clients on BroadcastMsg
	FilterCallback func(c *Client, data interface{}) bool

	// startWaitGroup is used to synchronize the start of the writePump and receivePong func
	startWaitGroup sync.WaitGroup

	// shutdownWaitGroup is used wait until writePump and receivePong func stopped
	shutdownWaitGroup sync.WaitGroup

	// indicates that the client was shut down
	shutdownFlag *atomic.Bool

	// indicates the max amount of bytes that will be read from a client, i.e. the max message size
	readLimit int64
}

// checkPong checks if the client is still available and answers to the ping messages
// that are sent periodically in the writePump function.
//
// at most one reader per websocket connection is allowed.
func (c *Client) checkPong() {

	defer func() {
		select {
		case <-c.hub.ctx.Done():
		case <-c.ExitSignal:
			// the Hub closed the channel.
		default:
			// send a unregister message to the hub
			c.hub.unregister <- c
		}

		if c.ReceiveChan != nil {
			// drain and close the receive channel
		drainLoop:
			for {
				select {
				case <-c.ReceiveChan:
				default:
					break drainLoop
				}
			}
			close(c.ReceiveChan)
		}

		c.shutdownWaitGroup.Done()
	}()

	c.startWaitGroup.Done()

	c.conn.SetReadLimit(c.readLimit)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.hub.logger.Warnf("Websocket SetReadDeadline error: %v", err)
	}

	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			c.hub.logger.Warnf("Websocket SetReadDeadline error: %v", err)
		}

		return nil
	})

	for {
		msgType, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Warnf("Websocket ReadMessage error: %v", err)
			}

			return
		}

		if c.ReceiveChan != nil {
			select {

			case <-c.hub.ctx.Done():
				return

			case <-c.ExitSignal:
				// the Hub closed the channel.
				return

			case c.ReceiveChan <- &WebsocketMsg{MsgType: msgType, Data: data}:
				// send the received message to the user.
			}
		}
	}
}

// writePump pumps messages from the node to the websocket connection.
//
// at most one writer per websocket connection is allowed.
func (c *Client) writePump() {

	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		// signal the hub to not send messages to sendChan anymore
		close(c.sendChanClosed)

		// mark the client as shutdown
		c.shutdownFlag.Store(true)

		// stop the ping ticker
		pingTicker.Stop()

		select {
		case <-c.hub.ctx.Done():
		case <-c.ExitSignal:
			// the Hub closed the channel.
		default:
			// send a unregister message to the hub
			c.hub.unregister <- c
		}

		// close the websocket connection
		c.conn.Close()

		c.shutdownWaitGroup.Done()
	}()

	c.startWaitGroup.Done()

	for {
		select {

		case <-c.hub.ctx.Done():
			return

		case <-c.ExitSignal:
			// the Hub closed the channel.
			if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
				c.hub.logger.Warnf("Websocket WriteMessage error: %v", err)
			}

			return

		case msg, ok := <-c.sendChan:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.hub.logger.Warnf("Websocket SetWriteDeadline error: %v", err)
			}

			if !ok {
				// the Hub closed the channel.
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					c.hub.logger.Warnf("Websocket WriteMessage error: %v", err)
				}

				return
			}

			if err := c.conn.WriteJSON(msg); err != nil {
				c.hub.logger.Warnf("Websocket error: %v", err)

				return
			}

		case <-pingTicker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.hub.logger.Warnf("Websocket SetWriteDeadline error: %v", err)
			}

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Send sends a message to the client.
func (c *Client) Send(msg interface{}, dontDrop ...bool) {
	if c.hub.shutdownFlag.Load() {
		// hub was already shut down
		return
	}

	if c.shutdownFlag.Load() {
		// client was already shutdown
		return
	}

	if len(dontDrop) > 0 && dontDrop[0] {
		select {
		case <-c.hub.ctx.Done():
		case <-c.ExitSignal:
		case <-c.sendChanClosed:
		case c.sendChan <- msg:
		}

		return
	}

	select {
	case <-c.hub.ctx.Done():
	case <-c.ExitSignal:
	case <-c.sendChanClosed:
	case c.sendChan <- msg:
	default:
	}
}

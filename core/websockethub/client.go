package websockethub

import (
	"context"
	"sync"
	"time"

	"go.uber.org/atomic"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)
 
const (
	// time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// WebsocketMsg is a message received via websocket.
type WebsocketMsg struct {
	// MsgType is the type of the message based on RFC 6455.
	MsgType websocket.MessageType
	// Data is the received data of the message.
	Data []byte
}

// ClientID is the ID of a client.
type ClientID uint32

// Client is a middleman between the node and the websocket connection.
type Client struct {
	// the id of the client.
	id ClientID

	// the websocket hub the client is connected to.
	hub *Hub

	// the websocket connection.
	conn *websocket.Conn

	// a context which is canceled when the ping times out and the client should be dropped.
	keepAliveContext context.Context
	keepAliveCancel  context.CancelFunc

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

	// onDisconnect gets called when the client was disconnected
	onDisconnect func(*Client)

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

// ID returns the id of the client.
func (c *Client) ID() ClientID {
	return c.id
}

func (c *Client) keepAlive() {
	ticker := time.NewTimer(time.Millisecond)

	defer c.keepAliveCancel()
	defer ticker.Stop()

	for {
		select {
		case <-c.hub.ctx.Done():
		case <-c.ExitSignal:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(c.hub.ctx, pongWait)
			err := c.conn.Ping(ctx)
			cancel()

			if err != nil {
				return
			}

			ticker.Reset(pingPeriod)
		}
	}
}

// readPump reads incoming messages and stops if the client does not respond to the keep alive pings
// that are sent periodically in the keepAlive function.
//
// at most one reader per websocket connection is allowed.
func (c *Client) readPump() {

	defer func() {
		select {
		case <-c.hub.ctx.Done():
		case <-c.ExitSignal:
			// the Hub closed the channel.
		default:
			// send an unregister message to the hub
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

	for {
		msgType, data, err := c.conn.Read(c.keepAliveContext)

		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusGoingAway || websocket.CloseStatus(err) == websocket.StatusAbnormalClosure {
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
			// send an unregister message to the hub
			c.hub.unregister <- c
		}

		// close the websocket connection
		if err := c.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			c.hub.logger.Warnf("Websocket closing error: %v", err)
		}

		c.shutdownWaitGroup.Done()
	}()

	c.startWaitGroup.Done()

	for {
		select {

		case <-c.hub.ctx.Done():
			return

		case <-c.ExitSignal:
			// the Hub closed the channel.
			if err := c.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
				c.hub.logger.Warnf("Websocket closing error: %v", err)
			}

			return

		case msg, ok := <-c.sendChan:
			if !ok {
				// the Hub closed the channel.
				if err := c.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
					c.hub.logger.Warnf("Websocket closing error: %v", err)
				}

				return
			}

			if err := wsjson.Write(c.keepAliveContext, c.conn, msg); err != nil {
				c.hub.logger.Warnf("Websocket error: %v", err)
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
		case <-c.keepAliveContext.Done():
		case <-c.hub.ctx.Done():
		case <-c.ExitSignal:
		case <-c.sendChanClosed:
		case c.sendChan <- msg:
		}

		return
	}

	select {
	case <-c.keepAliveContext.Done():
	case <-c.hub.ctx.Done():
	case <-c.ExitSignal:
	case <-c.sendChanClosed:
	case c.sendChan <- msg:
	default:
	}
}

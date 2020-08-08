package websockethub

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	// time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// maximum message size allowed from peer.
	maxMessageSize = 125 // 125 is the maximum payload size for ping pongs
)

// Client is a middleman between the node and the websocket connection.
type Client struct {
	hub *Hub

	// the websocket connection.
	conn *websocket.Conn

	// a channel which is closed when the websocket client is disconnected.
	ExitSignal chan struct{}

	// buffered channel of outbound messages.
	sendChan chan interface{}

	// onConnect gets called when the client was registered
	onConnect func(*Client)

	// FilterCallback is used to filter messages to clients on BroadcastMsg
	FilterCallback func(c *Client, data interface{}) bool
}

// checkPong checks if the client is still available and answers to the ping messages
// that are sent periodically in the writePump function.
//
// at most one reader per websocket connection is allowed
func (c *Client) checkPong() {

	defer func() {
		// send a unregister message to the hub
		c.hub.unregister <- c
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Warnf("Websocket error: %v", err)
			}
			return
		}
	}
}

// writePump pumps messages from the node to the websocket connection.
//
// at most one writer per websocket connection is allowed
func (c *Client) writePump() {

	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		// stop the ping ticker
		pingTicker.Stop()

		// send a unregister message to the hub
		c.hub.unregister <- c

		// close the websocket connection
		c.conn.Close()
	}()

	for {
		select {

		case <-c.hub.shutdownSignal:
			return

		case <-c.ExitSignal:
			// the Hub closed the channel.
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return

		case msg, ok := <-c.sendChan:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// the Hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(msg); err != nil {
				c.hub.logger.Warnf("Websocket error: %v", err)
				return
			}

		case <-pingTicker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Send sends a message to the client
func (c *Client) Send(msg interface{}, dontDrop ...bool) {
	if len(dontDrop) > 0 && dontDrop[0] {
		select {
		case <-c.hub.shutdownSignal:
		case <-c.ExitSignal:
		case c.sendChan <- msg:
		}
		return
	}

	select {
	case <-c.hub.shutdownSignal:
	case <-c.ExitSignal:
	case c.sendChan <- msg:
	default:
	}
}

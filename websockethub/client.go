package websockethub

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 125 // 125 is the maximum payload size for ping pongs

	// Maximum size of queued messages that should be sent to the peer.
	sendChannelSize = 100
)

// Client is a middleman between the node and the websocket connection.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	sendChan chan interface{}
}

// checkPong checks if the client is still available and answers to the ping messages
// that are sent periodically in the writePump function.
//
// At most one reader per websocket connection is allowed
func (c *Client) checkPong() {

	defer func() {
		// Send a unregister message to the hub
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
// At most one writer per websocket connection is allowed
func (c *Client) writePump() {

	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		// stop the ping ticker
		pingTicker.Stop()

		// Send a unregister message to the hub
		c.hub.unregister <- c

		// close the websocket connection
		c.conn.Close()
	}()

	for {
		select {

		case <-c.hub.shutdownSignal:
			return

		case msg, ok := <-c.sendChan:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The Hub closed the channel.
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
func (c *Client) Send(msg interface{}) {
	select {
	case c.sendChan <- msg:
	default:
	}
}

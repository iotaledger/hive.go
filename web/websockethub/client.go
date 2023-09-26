package websockethub

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/izuc/zipp.foundation/logger"
	"github.com/izuc/zipp.foundation/runtime/timeutil"
)

const (
	// time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// time allowed to read the next pong message from the peer.
	pongWait = 5 * time.Second

	// send pings to peer with this period.
	pingPeriod = 30 * time.Second
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
	*logger.WrappedLogger

	// the id of the client.
	id ClientID

	// the websocket hub the client is connected to.
	hub *Hub

	// the websocket connection.
	conn *websocket.Conn

	// a context which is canceled when the ping times out and the client should be dropped.
	ctx    context.Context
	cancel context.CancelFunc

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
	shutdownFlag atomic.Bool

	// indicates the max amount of bytes that will be read from a client, i.e. the max message size
	readLimit int64
}

func NewClient(hub *Hub, conn *websocket.Conn, onConnect func(client *Client), onDisconnect func(client *Client)) *Client {
	ctx, cancel := context.WithCancel(hub.ctx)

	clientID := ClientID(hub.lastClientID.Add(1))

	return &Client{
		WrappedLogger:  logger.NewWrappedLogger(hub.logger.Named(fmt.Sprintf("client %d", clientID))),
		id:             clientID,
		hub:            hub,
		conn:           conn,
		ctx:            ctx,
		cancel:         cancel,
		ExitSignal:     make(chan struct{}),
		sendChan:       make(chan interface{}, hub.clientSendChannelSize),
		sendChanClosed: make(chan struct{}),
		onConnect:      onConnect,
		onDisconnect:   onDisconnect,
		readLimit:      hub.clientReadLimit,
	}
}

// ID returns the id of the client.
func (c *Client) ID() ClientID {
	return c.id
}

// Context returns the client context which is canceled when the ping times out and the client should be dropped.
func (c *Client) Context() context.Context {
	return c.ctx
}

// keepAlive sends ping messages to the client and waits for pong responses.
// if no pong response is received in time, the client context is canceled.
func (c *Client) keepAlive() {
	// start the timer with 0, so it fires immediately
	pingTimer := time.NewTimer(0)

	defer func() {
		timeutil.CleanupTimer(pingTimer)

		// always cancel the client context if we exit this function to clean up the client
		c.cancel()

		c.shutdownWaitGroup.Done()
	}()

	sendPing := func() error {
		pongCtx, pongCancel := context.WithTimeout(c.ctx, pongWait)
		defer pongCancel()

		if err := c.conn.Ping(pongCtx); err != nil {
			return err
		}

		// reset the ping timer, so a new ping event is fired after "pingPeriod".
		// we can safely reset the timer here, because the timer channel was consumed already.
		pingTimer.Reset(pingPeriod)

		return nil
	}

	c.startWaitGroup.Done()

	for {
		// we need to nest the pingTimer.C into the default case because
		// the select cases are executed in random order if multiple
		// conditions are true at the time of entry in the select case.
		select {
		case <-c.ctx.Done():
			// the client context is done
			return
		case <-c.ExitSignal:
			// the Hub closed the channel
			return
		default:
			select {
			case <-c.ctx.Done():
				// the client context is done
				return
			case <-c.ExitSignal:
				// the Hub closed the channel
				return
			case <-pingTimer.C:
				if err := sendPing(); err != nil {
					// failed to send ping or receive pong
					// => client seems to be unhealthy
					c.LogWarn(err.Error())

					// send an unregister message to the hub
					_ = c.hub.Unregister(c)

					return
				}
			}
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
		case <-c.ctx.Done():
			// the client context is done
		case <-c.ExitSignal:
			// the Hub closed the channel
		default:
			// send an unregister message to the hub
			_ = c.hub.Unregister(c)
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
		msgType, data, err := c.conn.Read(c.ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusGoingAway || websocket.CloseStatus(err) == websocket.StatusAbnormalClosure {
				c.LogWarnf("Websocket ReadMessage error: %v", err)
			}

			return
		}

		if c.ReceiveChan != nil {
			// we need to nest the ReceiveChan into the default case because
			// the select cases are executed in random order if multiple
			// conditions are true at the time of entry in the select case.
			select {
			case <-c.ctx.Done():
				// the client context is done
				return
			case <-c.ExitSignal:
				// the Hub closed the channel
				return
			default:
				select {
				case <-c.ctx.Done():
					// the client context is done
					return
				case <-c.ExitSignal:
					// the Hub closed the channel
					return
				case c.ReceiveChan <- &WebsocketMsg{MsgType: msgType, Data: data}:
					// send the received message to the user
				}
			}
		}
	}
}

// writePump pumps messages from the node to the websocket connection.
//
// at most one writer per websocket connection is allowed.
func (c *Client) writePump() {
	defer func() {
		// signal the hub to not send messages to sendChan anymore
		close(c.sendChanClosed)

		// mark the client as shutdown
		c.shutdownFlag.Store(true)

		select {
		case <-c.ctx.Done():
			// the client context is done
		case <-c.ExitSignal:
			// the Hub closed the channel
		default:
			// send an unregister message to the hub
			_ = c.hub.Unregister(c)
		}

		// close the websocket connection
		if err := c.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			c.LogWarnf("Websocket closing error: %v", err)
		}

		c.shutdownWaitGroup.Done()
	}()

	sendMsg := func(msg any) error {
		ctx, cancel := context.WithTimeout(c.ctx, writeWait)
		defer cancel()

		return wsjson.Write(ctx, c.conn, msg)
	}

	closeConnection := func() {
		if err := c.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			c.LogWarnf("Websocket closing error: %v", err)
		}
	}

	for {
		// we need to nest the c.sendChan into the default case because
		// the select cases are executed in random order if multiple
		// conditions are true at the time of entry in the select case.
		select {
		case <-c.ctx.Done():
			// the client context is done
			return
		case <-c.ExitSignal:
			// the Hub closed the channel
			closeConnection()
			return
		default:
			select {
			case <-c.ctx.Done():
				// the client context is done
				return
			case <-c.ExitSignal:
				// the Hub closed the channel
				closeConnection()
				return
			case msg, ok := <-c.sendChan:
				if !ok {
					// the Hub closed the channel
					closeConnection()
					return
				}

				if err := sendMsg(msg); err != nil {
					c.LogWarnf("Websocket error: %v", err)
					return
				}
			}
		}
	}
}

// Send sends a message to the client.
func (c *Client) Send(ctx context.Context, msg interface{}, dontDrop ...bool) error {
	if c.hub.Stopped() {
		// hub was already shut down
		return ErrWebsocketServerUnavailable
	}

	if c.shutdownFlag.Load() {
		// client was already shutdown
		return ErrClientDisconnected
	}

	if len(dontDrop) > 0 && dontDrop[0] {
		// we need to nest the sendChan into the default case because
		// the select cases are executed in random order if multiple
		// conditions are true at the time of entry in the select case.
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		case <-ctx.Done():
			return ctx.Err()
		case <-c.ExitSignal:
			return ErrClientDisconnected
		case <-c.sendChanClosed:
			return ErrClientDisconnected
		default:
			select {
			case <-c.ctx.Done():
				return c.ctx.Err()
			case <-ctx.Done():
				return ctx.Err()
			case <-c.ExitSignal:
				return ErrClientDisconnected
			case <-c.sendChanClosed:
				return ErrClientDisconnected
			case c.sendChan <- msg:
				return nil
			}
		}
	}

	// we need to nest the sendChan into the default case because
	// the select cases are executed in random order if multiple
	// conditions are true at the time of entry in the select case.
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ExitSignal:
		return ErrClientDisconnected
	case <-c.sendChanClosed:
		return ErrClientDisconnected
	default:
		select {
		case c.sendChan <- msg:
			return nil
		default:
			return nil
		}
	}
}

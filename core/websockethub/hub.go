package websockethub

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/atomic"

	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/hive.go/core/logger"
)

type ClientConnectionEvent struct {
	ID ClientID
}

// Events contains all the events that are triggered by the websocket hub.
type Events struct {
	// A ClientConnected event is triggered, when a new client has connected to the websocket hub.
	ClientConnected *event.Event[*ClientConnectionEvent]
	// A ClientDisconnected event is triggered, when a client has disconnected from the websocket hub.
	ClientDisconnected *event.Event[*ClientConnectionEvent]
}

func newEvents() *Events {
	return &Events{
		ClientConnected:    event.New[*ClientConnectionEvent](),
		ClientDisconnected: event.New[*ClientConnectionEvent](),
	}
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {

	// websocket Upgrader.
	upgrader *websocket.Upgrader

	// used Logger instance.
	logger *logger.Logger

	// registered clients.
	clients map[*Client]struct{}

	// maximum size of queued messages that should be sent to the peer.
	clientSendChannelSize int

	// inbound messages from the clients.
	broadcast chan *message

	// register requests from the clients.
	register chan *Client

	// unregister requests from clients.
	unregister chan *Client

	// context of the websocket hub
	ctx       context.Context
	ctxCancel context.CancelFunc

	// indicates that the websocket hub was shut down
	shutdownFlag *atomic.Bool

	// indicates the max amount of bytes that will be read from a client, i.e. the max message size
	clientReadLimit int64

	// lastClientID holds the ClientID of the last connected client
	lastClientID *atomic.Uint32

	// events of the websocket hub
	events *Events
}

// message is a message that is sent to the broadcast channel.
type message struct {
	data     interface{}
	dontDrop bool
}

func NewHub(logger *logger.Logger, upgrader *websocket.Upgrader, broadcastQueueSize int, clientSendChannelSize int, clientReadLimit int64) *Hub {
	ctx, ctxCancel := context.WithCancel(context.Background())

	return &Hub{
		logger:                logger,
		upgrader:              upgrader,
		clientSendChannelSize: clientSendChannelSize,
		clients:               make(map[*Client]struct{}),
		broadcast:             make(chan *message, broadcastQueueSize),
		register:              make(chan *Client, 1),
		unregister:            make(chan *Client, 1),
		ctx:                   ctx,
		ctxCancel:             ctxCancel,
		shutdownFlag:          atomic.NewBool(false),
		clientReadLimit:       clientReadLimit,
		lastClientID:          atomic.NewUint32(0),
		events:                newEvents(),
	}
}

// Events returns all the events that are triggered by the websocket hub.
func (h *Hub) Events() *Events {
	return h.events
}

// BroadcastMsg sends a message to all clients.
func (h *Hub) BroadcastMsg(data interface{}, dontDrop ...bool) {
	if h.shutdownFlag.Load() {
		// hub was already shut down
		return
	}

	notDrop := false
	if len(dontDrop) > 0 {
		notDrop = dontDrop[0]
	}

	msg := &message{data: data, dontDrop: notDrop}

	if notDrop {
		select {
		case <-h.ctx.Done():
		case h.broadcast <- msg:
		}

		return
	}

	select {
	case <-h.ctx.Done():
	case h.broadcast <- msg:
	default:
	}
}

func (h *Hub) removeClient(client *Client) {
	delete(h.clients, client)
	close(client.ExitSignal)

	// wait until writePump and checkPong finished
	client.shutdownWaitGroup.Wait()

	// drain the send channel
drainLoop:
	for {
		select {
		case <-client.sendChan:
		default:
			break drainLoop
		}
	}

	if client.onDisconnect != nil {
		client.onDisconnect(client)
	}
	h.events.ClientDisconnected.Trigger(&ClientConnectionEvent{ID: client.id})

	// We do not call "close(client.sendChan)" because we have multiple senders.
	//
	// As written at https://go101.org/article/channel-closing.html
	// A channel will be eventually garbage collected if no goroutines reference it any more,
	// whether it is closed or not.
	// So the gracefulness of closing a channel here is not to close the channel.
}

// Returns the number of websocket clients.
func (h *Hub) Clients() int {
	return len(h.clients)
}

// Run starts the hub.
func (h *Hub) Run(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			h.shutdownFlag.Store(true)
			h.ctxCancel()

			for client := range h.clients {
				h.removeClient(client)
			}

			return

		case client := <-h.register:
			// register client
			h.clients[client] = struct{}{}

			client.shutdownWaitGroup.Add(2)

			// first start the write pump to answer requests from checkPong
			client.startWaitGroup.Add(1)
			go client.writePump()
			client.startWaitGroup.Wait()

			// wait until checkPong started, before calling onConnect
			client.startWaitGroup.Add(1)
			go client.checkPong()
			client.startWaitGroup.Wait()

			if client.onConnect != nil {
				client.onConnect(client)
			}
			h.events.ClientConnected.Trigger(&ClientConnectionEvent{ID: client.id})

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.removeClient(client)
				h.logger.Infof("Removed websocket client")
			}

		case message := <-h.broadcast:
			if message.dontDrop {
				for client := range h.clients {
					if client.FilterCallback != nil {
						if !client.FilterCallback(client, message.data) {
							// do not broadcast the message to this client
							continue
						}
					}

					select {
					case <-ctx.Done():
					case <-client.ExitSignal:
					case <-client.sendChanClosed:
					case client.sendChan <- message.data:
					}
				}

				continue
			}
			for client := range h.clients {
				if client.FilterCallback != nil {
					if !client.FilterCallback(client, message.data) {
						// do not broadcast the message to this client
						continue
					}
				}

				select {
				case <-ctx.Done():
				case <-client.ExitSignal:
				case <-client.sendChanClosed:
				case client.sendChan <- message.data:
				default:
				}
			}
		}
	}
}

// ServeWebsocket handles websocket requests from the peer.
// onCreate gets called when the client is created.
// onConnect gets called when the client was registered.
func (h *Hub) ServeWebsocket(w http.ResponseWriter,
	r *http.Request,
	onCreate func(client *Client),
	onConnect func(client *Client),
	onDisconnect func(client *Client)) {

	if h.shutdownFlag.Load() {
		// hub was already shut down
		return
	}

	defer func() {
		if r := recover(); r != nil {
			h.logger.Errorf("recovered from ServeWebsocket func: %s", r)
		}
	}()

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Warnf("upgrade websocket error: %v", err)

		return
	}
	conn.EnableWriteCompression(true)

	clientID := ClientID(h.lastClientID.Inc())

	client := &Client{
		id:             clientID,
		hub:            h,
		conn:           conn,
		ExitSignal:     make(chan struct{}),
		sendChan:       make(chan interface{}, h.clientSendChannelSize),
		sendChanClosed: make(chan struct{}),
		onConnect:      onConnect,
		onDisconnect:   onDisconnect,
		shutdownFlag:   atomic.NewBool(false),
		readLimit:      h.clientReadLimit,
	}

	if onCreate != nil {
		onCreate(client)
	}

	select {
	case <-h.ctx.Done():
	case h.register <- client:
	}
}

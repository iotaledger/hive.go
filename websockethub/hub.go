package websockethub

import (
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/iotaledger/hive.go/logger"
)

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

	// signal shutdown of the websocket hub
	shutdownSignal chan struct{}

	// indicates that the websocket hub was shut down
	shutdownFlag bool

	// indicates the max amount of bytes that will be read from a client, i.e. the max message size
	clientReadLimit int64
}

// message is a message that is sent to the broadcast channel.
type message struct {
	data     interface{}
	dontDrop bool
}

func NewHub(logger *logger.Logger, upgrader *websocket.Upgrader, broadcastQueueSize int, clientSendChannelSize int, clientReadLimit int64) *Hub {
	return &Hub{
		logger:                logger,
		upgrader:              upgrader,
		clientSendChannelSize: clientSendChannelSize,
		clients:               make(map[*Client]struct{}),
		broadcast:             make(chan *message, broadcastQueueSize),
		register:              make(chan *Client, 1),
		unregister:            make(chan *Client, 1),
		shutdownSignal:        make(chan struct{}),
		clientReadLimit:       clientReadLimit,
	}
}

// BroadcastMsg sends a message to all clients.
func (h *Hub) BroadcastMsg(data interface{}, dontDrop ...bool) {
	if h.shutdownFlag {
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
		case <-h.shutdownSignal:
		case h.broadcast <- msg:
		}
		return
	}

	select {
	case <-h.shutdownSignal:
	case h.broadcast <- msg:
	default:
	}
}

// Run starts the hub.
func (h *Hub) Run(shutdownSignal <-chan struct{}) {

	for {
		select {
		case <-shutdownSignal:
			h.shutdownFlag = true
			close(h.shutdownSignal)

			for client := range h.clients {
				delete(h.clients, client)
				close(client.ExitSignal)

				// wait until writePump and checkPong finished
				client.shutdownWaitGroup.Wait()

				if client.ReceiveChan != nil {
					close(client.ReceiveChan)
				}
				close(client.sendChan)
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

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.ExitSignal)

				// wait until writePump and checkPong finished
				client.shutdownWaitGroup.Wait()

				if client.ReceiveChan != nil {
					close(client.ReceiveChan)
				}
				close(client.sendChan)
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
					case <-shutdownSignal:
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
				case <-shutdownSignal:
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
func (h *Hub) ServeWebsocket(w http.ResponseWriter, r *http.Request, onCreate func(client *Client), onConnect func(client *Client)) {
	if h.shutdownFlag {
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

	client := &Client{
		hub:            h,
		conn:           conn,
		ExitSignal:     make(chan struct{}),
		sendChan:       make(chan interface{}, h.clientSendChannelSize),
		sendChanClosed: make(chan struct{}),
		onConnect:      onConnect,
		readLimit:      h.clientReadLimit,
	}

	if onCreate != nil {
		onCreate(client)
	}

	select {
	case <-h.shutdownSignal:
	case h.register <- client:
	}
}

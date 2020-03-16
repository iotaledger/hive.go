package websockethub

import (
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/iotaledger/hive.go/logger"
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {

	// Websocket Upgrader.
	upgrader *websocket.Upgrader

	// Used Logger instance.
	logger *logger.Logger

	// Registered clients.
	clients map[*Client]struct{}

	// Inbound messages from the clients.
	broadcast chan interface{}

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	shutdownSignal <-chan struct{}
}

func NewHub(logger *logger.Logger, upgrader *websocket.Upgrader, broadcastQueueSize int) *Hub {
	return &Hub{
		logger:     logger,
		upgrader:   upgrader,
		clients:    make(map[*Client]struct{}),
		broadcast:  make(chan interface{}, broadcastQueueSize),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// BroadcastMsg sends a message to all clients
func (h *Hub) BroadcastMsg(msg interface{}) {
	select {
	case h.broadcast <- msg:
	default:
	}
}

// Start the hub
func (h *Hub) Run(shutdownSignal <-chan struct{}) {

	for {
		select {
		case <-shutdownSignal:
			for client := range h.clients {
				delete(h.clients, client)
				close(client.sendChan)
			}
			return

		case client := <-h.register:
			// register client
			h.clients[client] = struct{}{}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.sendChan)
				h.logger.Infof("Removed websocket client")
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.sendChan <- message:
				default:
				}
			}
		}
	}
}

// ServeWebsocket handles websocket requests from the peer.
func (h *Hub) ServeWebsocket(w http.ResponseWriter, r *http.Request, onConnect ...func(client *Client)) {

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Warnf("Upgrade websocket: %v", err)
		return
	}

	client := &Client{hub: h, conn: conn, sendChan: make(chan interface{}, sendChannelSize)}
	h.register <- client

	go client.checkPong()
	go client.writePump()

	if len(onConnect) > 0 {
		onConnect[0](client)
	}
}

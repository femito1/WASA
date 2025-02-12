// File: service/api/realtime.go
package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

// EventMessage represents a realtime event.
type EventMessage struct {
	Type           string      `json:"type"`
	ConversationId uint64      `json:"conversationId,omitempty"`
	Payload        interface{} `json:"payload,omitempty"`
	Recipients     []uint64    `json:"recipients,omitempty"`
}

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *EventMessage
	mu         sync.Mutex
}

// NewHub creates a new realtime Hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *EventMessage),
	}
}

// Run starts the hub event loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case event := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				// If Recipients is set, filter the clients.
				if len(event.Recipients) > 0 {
					shouldSend := false
					for _, id := range event.Recipients {
						if id == client.userID {
							shouldSend = true
							break
						}
					}
					if !shouldSend {
						continue
					}
				}
				message, err := json.Marshal(event)
				if err != nil {
					continue
				}
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client represents a websocket connection from a user.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID uint64
}

// readPump reads messages from the websocket.
// (In this application we do not expect incoming messages from the client.)
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

// writePump sends messages to the websocket.
func (c *Client) writePump() {
	defer c.conn.Close()
	for range c.send {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}

// serveWs upgrades the HTTP connection to a WebSocket and registers the client.
func (rt *_router) serveWs(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get token from query parameter.
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token required", http.StatusUnauthorized)
		return
	}
	userID, err := ExtractUserIDFromTokenString(token)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	client := &Client{
		hub:    rt.hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
	}
	rt.hub.register <- client
	go client.writePump()
	go client.readPump()
}

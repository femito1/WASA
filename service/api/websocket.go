package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/femito1/WASA/service/api/reqcontext"
	"github.com/femito1/WASA/service/database"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // In development; in production, implement proper origin checking
	},
}

type WSClient struct {
	conn     *websocket.Conn
	userId   uint64
	mu       sync.Mutex
	channels map[uint64]bool // Track which conversation channels this client is subscribed to
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

var (
	clients   = make(map[uint64]*WSClient)
	clientsMu sync.RWMutex
)

// Add this to your _router struct handlers initialization
func (rt *_router) setupWebSocket() {
	rt.router.Handle("GET", "/users/:id/ws", rt.wrap(rt.handleWebSocket))
}

func (rt *_router) handleWebSocket(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Get user ID from URL parameters (consistent with other handlers)
	userID, err := strconv.ParseUint(ps.ByName("id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("invalid user ID")
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Upgrade the HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ctx.Logger.WithError(err).Error("websocket upgrade failed")
		return
	}

	defer func() {
		if err := conn.Close(); err != nil {
			ctx.Logger.WithError(err).Error("error closing websocket connection")
		}
	}()

	client := &WSClient{
		conn:     conn,
		userId:   userID,
		channels: make(map[uint64]bool),
	}

	// Register client
	clientsMu.Lock()
	clients[client.userId] = client
	clientsMu.Unlock()

	// Cleanup on disconnect
	defer func() {
		clientsMu.Lock()
		delete(clients, client.userId)
		clientsMu.Unlock()
	}()

	// Handle incoming messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				ctx.Logger.WithError(err).Error("websocket read error")
			}
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Handle subscription messages
		if msg.Type == "subscribe" {
			if convId, ok := msg.Payload.(float64); ok {
				client.channels[uint64(convId)] = true
			}
		}
	}
}

// Send a message to all clients in a conversation
func (rt *_router) broadcastToConversation(convId uint64, msgType string, payload interface{}) {
	message := WSMessage{
		Type:    msgType,
		Payload: payload,
	}

	data, err := json.Marshal(message)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to marshal websocket message")
		return
	}

	// Get conversation members from the database
	conv, err := rt.db.GetConversation(0, convId, nil)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to get conversation")
		return
	}

	clientsMu.RLock()
	defer clientsMu.RUnlock()

	for _, member := range conv.Members {
		if client, ok := clients[member.Id]; ok {
			client.mu.Lock()
			if err := client.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				rt.baseLogger.WithError(err).Error("failed to send websocket message")
			}
			client.mu.Unlock()
		}
	}
}

// Call this when a new message is created
func (rt *_router) notifyNewMessage(convId uint64, message database.Message) {
	rt.broadcastToConversation(convId, "new_message", message)
}

// Call this when a message status changes
func (rt *_router) notifyMessageUpdate(convId uint64, messageId uint64, status string) {
	rt.broadcastToConversation(convId, "message_update", map[string]interface{}{
		"conversationId": convId,
		"messageId":      messageId,
		"state":          status,
	})
}

// Call this when a message is read
func (rt *_router) notifyMessageRead(convId uint64, messageId uint64, userId uint64) {
	rt.broadcastToConversation(convId, "message_read", map[string]interface{}{
		"conversationId": convId,
		"messageId":      messageId,
		"userId":         userId,
		"timestamp":      time.Now().Format(time.RFC3339),
	})
}

// notifyProfileUpdate broadcasts profile picture updates to relevant conversations
func (rt *_router) notifyProfileUpdate(userId uint64, newPicture string) {
	// Get all conversations where this user is a member
	conversations, err := rt.db.GetConversations(userId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to get user conversations for profile update")
		return
	}

	update := WSMessage{
		Type: "profile_update",
		Payload: map[string]interface{}{
			"userId":         userId,
			"profilePicture": newPicture,
		},
	}

	// Broadcast to each conversation
	for _, conv := range conversations {
		rt.broadcastToConversation(conv.Id, update.Type, update.Payload)
	}
}

// Add to _router struct
func (rt *_router) notifyMessageReaction(convId uint64, messageId uint64, reaction database.Reaction) {
	rt.broadcastToConversation(convId, "message_reaction", map[string]interface{}{
		"conversationId": convId,
		"messageId":      messageId,
		"reaction":       reaction,
	})
}

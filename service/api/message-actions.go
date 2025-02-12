package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/femito1/WASA/service/api/reqcontext"
	"github.com/femito1/WASA/service/database"
	"github.com/julienschmidt/httprouter"
)

// In service/api/message-actions.go, function sendMessage:
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userIdStr := ps.ByName("id")
	convIdStr := ps.ByName("convId")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	// Extract the user ID from the bearer token.
	tokenUserID, err := ExtractUserIDFromToken(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Ensure that the token's user ID matches the URL's user ID.
	if tokenUserID != userId {
		http.Error(w, "forbidden: you cannot update another user's details", http.StatusForbidden)
		return
	}

	convId, err := strconv.ParseUint(convIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid conversation id", http.StatusBadRequest)
		return
	}
	var reqPayload struct {
		Content string  `json:"content"`
		Format  string  `json:"format"`
		ReplyTo *uint64 `json:"replyTo,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Verify that the sender exists.
	sender, err := rt.db.CheckUserById(database.User{Id: userId})
	if err != nil {
		http.Error(w, "sender not found", http.StatusNotFound)
		return
	}

	msg, err := rt.db.CreateMessage(sender, convId, reqPayload.Content, reqPayload.Format, reqPayload.ReplyTo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode sendMessage response")
	}
}

// deleteMessage handles DELETE /users/:id/conversations/:convId/messages/:msgId.
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userIdStr := ps.ByName("id")
	convIdStr := ps.ByName("convId")
	msgIdStr := ps.ByName("msgId")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	// Extract the user ID from the bearer token.
	tokenUserID, err := ExtractUserIDFromToken(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Ensure that the token's user ID matches the URL's user ID.
	if tokenUserID != userId {
		http.Error(w, "forbidden: you cannot update another user's details", http.StatusForbidden)
		return
	}

	convId, err := strconv.ParseUint(convIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid conversation id", http.StatusBadRequest)
		return
	}
	msgId, err := strconv.ParseUint(msgIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid message id", http.StatusBadRequest)
		return
	}
	// Verify the user exists.
	user, err := rt.db.CheckUserById(database.User{Id: userId})
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if err := rt.db.DeleteMessage(user, convId, msgId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// forwardMessage handles POST /users/:id/conversations/:convId/messages/:msgId/forward.
// It expects a JSON payload with { "targetConversationId": number }.
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userIdStr := ps.ByName("id")
	convIdStr := ps.ByName("convId")
	msgIdStr := ps.ByName("msgId")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	// Extract and validate convId and msgId.
	convId, err := strconv.ParseUint(convIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid conversation id", http.StatusBadRequest)
		return
	}
	msgId, err := strconv.ParseUint(msgIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid message id", http.StatusBadRequest)
		return
	}
	var reqPayload struct {
		TargetConversationId uint64 `json:"targetConversationId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Verify that the user exists.
	user, err := rt.db.CheckUserById(database.User{Id: userId})
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Verify that the target conversation exists and that the user is a member.
	_, err = rt.db.GetConversation(userId, reqPayload.TargetConversationId, nil)
	if err != nil {
		http.Error(w, "target conversation not found or access denied", http.StatusBadRequest)
		return
	}

	msg, err := rt.db.ForwardMessage(user, convId, msgId, reqPayload.TargetConversationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode forwardMessage response")
	}
}

// reactToMessage handles POST /users/:id/conversations/:convId/messages/:msgId/reaction.
// It expects a JSON payload with { "emoji": string }.
func (rt *_router) reactToMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Parse IDs from URL
	userIdStr := ps.ByName("id")
	convIdStr := ps.ByName("convId")
	msgIdStr := ps.ByName("msgId")

	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	convId, err := strconv.ParseUint(convIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid conversation id", http.StatusBadRequest)
		return
	}
	msgId, err := strconv.ParseUint(msgIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid message id", http.StatusBadRequest)
		return
	}

	// Verify the token's user id matches the URL's user id.
	tokenUserID, err := ExtractUserIDFromToken(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if tokenUserID != userId {
		http.Error(w, "forbidden: you cannot update another user's details", http.StatusForbidden)
		return
	}

	// Decode the reaction payload.
	var reqPayload struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if reqPayload.Emoji == "" {
		http.Error(w, "emoji is required", http.StatusBadRequest)
		return
	}

	// Verify the user exists.
	user, err := rt.db.CheckUserById(database.User{Id: userId})
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Call the database function to add the reaction.
	if err := rt.db.ReactToMessage(user, convId, msgId, reqPayload.Emoji); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with no content.
	w.WriteHeader(http.StatusNoContent)
}

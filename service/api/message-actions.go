package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/femito1/WASA/service/api/reqcontext"
	"github.com/femito1/WASA/service/database"
	"github.com/julienschmidt/httprouter"
)

// sendMessage handles POST /users/:id/conversations/:convId/messages.
// It expects a JSON payload with { "content": string, "format": string } and returns the created message.
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
		Content string `json:"content"`
		Format  string `json:"format"`
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
	msg, err := rt.db.CreateMessage(sender, convId, reqPayload.Content, reqPayload.Format)
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
	if err := rt.db.ForwardMessage(user, convId, msgId, reqPayload.TargetConversationId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

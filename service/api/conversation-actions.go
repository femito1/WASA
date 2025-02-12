package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/femito1/WASA/service/api/reqcontext"
	"github.com/femito1/WASA/service/database"
	"github.com/julienschmidt/httprouter"
)

// createGroup handles POST /users/:id/conversations.
// It accepts an optional JSON payload with "name" and "members" (an array of user IDs) and creates a new conversation.
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Parse the creator's user id from the URL.
	userIdStr := ps.ByName("id")
	creatorId, err := strconv.ParseUint(userIdStr, 10, 64)
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
	if tokenUserID != creatorId {
		http.Error(w, "forbidden: you cannot update another user's details", http.StatusForbidden)
		return
	}

	// Verify the creator exists.
	creator, err := rt.db.CheckUserById(database.User{Id: creatorId})
	if err != nil {
		http.Error(w, "creator not found", http.StatusNotFound)
		return
	}
	// Decode the (optional) payload.
	var reqPayload struct {
		Name    string   `json:"name"`
		Members []uint64 `json:"members"`
	}
	// Payload is optional.
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&reqPayload)
	}
	conv, err := rt.db.CreateConversation(creator, reqPayload.Name, reqPayload.Members)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(conv); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode createGroup response")
	}
}

// getMyConversations handles GET /users/:id/conversations.
// It returns all conversations in which the user (from URL) is a member.
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userIdStr := ps.ByName("id")
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

	convs, err := rt.db.GetConversations(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(convs); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode getMyConversations response")
	}
}

// getConversation handles GET /users/:id/conversations/:convId.
// Optionally, a query parameter "conversationName" may be provided for filtering.
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userIdStr := ps.ByName("id")
	convIdStr := ps.ByName("convId")

	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	convID, err := strconv.ParseUint(convIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid conversation id", http.StatusBadRequest)
		return
	}

	tokenUserID, err := ExtractUserIDFromToken(r)
	if err != nil || tokenUserID != userID {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Mark messages as read for the current user (i.e. messages not sent by this user)
	currentUser, err := rt.db.CheckUserById(database.User{Id: userID})
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if err = rt.db.MarkMessagesAsRead(currentUser, convID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	conv, err := rt.db.GetConversation(userID, convID, nil)
	if err != nil {
		if err.Error() == "user is not a member of this conversation" {
			http.Error(w, "forbidden: user is not a member of this conversation", http.StatusForbidden)
			return
		}
		if errors.Is(err, database.ErrConversationNotFound) {
			http.Error(w, "conversation not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(conv); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode getConversation response")
	}
}

// addtoGroup handles POST /users/:id/conversations/:convId/members.
// It expects a JSON payload with { "userIdToAdd": number }.
func (rt *_router) addtoGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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
		UserIdToAdd uint64 `json:"userIdToAdd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	conv, err := rt.db.AddUserToConversation(userId, convId, reqPayload.UserIdToAdd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(conv); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode addtoGroup response")
	}
}

// leaveGroup handles DELETE /users/:id/conversations/:convId/members.
// The user (from URL) is removed from the conversation.
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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
	if err := rt.db.RemoveUserFromConversation(userId, convId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// setGroupName handles PUT /users/:id/conversations/:convId/name.
// It expects a JSON payload with { "newName": string } and returns the updated conversation.
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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
		NewName string `json:"newName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	conv, err := rt.db.SetConversationName(userId, convId, reqPayload.NewName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(conv); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode setGroupName response")
	}
}

// setGroupPhoto handles PUT /users/:id/conversations/:convId/photo.
// It expects a JSON payload with { "newPhoto": string } and returns the updated conversation.
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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
		NewPhoto string `json:"newPhoto"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	conv, err := rt.db.SetConversationPhoto(userId, convId, reqPayload.NewPhoto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(conv); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode setGroupPhoto response")
	}
}

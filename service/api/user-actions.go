package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/femito1/WASA/service/api/reqcontext"
	"github.com/femito1/WASA/service/database"
	"github.com/julienschmidt/httprouter"
)

// doLogin handles POST /session.
// It decodes a JSON payload with { "name": string }, creates (or finds) the user in the database,
// and returns { "identifier": string } where the identifier is the user's unique ID.
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create the user (or return the existing user)
	dbUser, err := rt.db.CreateUser(database.User{
		Username: req.Name,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate a JWT token for the user.
	token, err := GenerateToken(dbUser.Id)
	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"identifier": token,
		"userId":     dbUser.Id,
	}); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode doLogin response")
	}

}

// listUsers handles GET /users.
// It accepts an optional query parameter "name" for filtering and returns an array of users.
func (rt *_router) listUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	nameFilter := r.URL.Query().Get("name")
	dbUsers, err := rt.db.ListUsers(nameFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert database users to API users.
	apiUsers := make([]User, len(dbUsers))
	for i, u := range dbUsers {
		var user User
		user.FromDatabase(u)
		apiUsers[i] = user
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiUsers); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode listUsers response")
	}
}

// setMyUserName handles PUT /users/:id.
// It expects a JSON payload with { "newName": string } and updates the user's username.
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var req struct {
		NewName string `json:"newName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Extract the user ID from the path parameter.
	idStr := ps.ByName("id")
	userId, err := strconv.ParseUint(idStr, 10, 64)
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

	// Fetch the current user from the database.
	dbUser, err := rt.db.CheckUserById(database.User{Id: userId})
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Update the username.
	updatedUser, err := rt.db.SetUsername(dbUser, req.NewName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated user.
	var user User
	user.FromDatabase(updatedUser)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode setMyUserName response")
	}
}

// setMyPhoto handles PUT /users/:id/photo.
// It expects a JSON payload with { "newPic": string } and updates the user's profile picture.
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var req struct {
		NewPic string `json:"newPic"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Extract the user ID from the path parameter.
	idStr := ps.ByName("id")
	userId, err := strconv.ParseUint(idStr, 10, 64)
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

	// Fetch the current user from the database.
	dbUser, err := rt.db.CheckUserById(database.User{Id: userId})
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Update the profile picture.
	updatedUser, err := rt.db.SetPhoto(dbUser, req.NewPic)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated user.
	var user User
	user.FromDatabase(updatedUser)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode setMyPhoto response")
	}
}

package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/femito1/WASA/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// addContact handles POST /users/:id/contacts.
// It expects a JSON payload with { "contactId": number }.
func (rt *_router) addContact(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userIdStr := ps.ByName("id")
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	// Validate token: ensure the user making the request is the same as in URL.
	tokenUserID, err := ExtractUserIDFromToken(r)
	if err != nil || tokenUserID != userID {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var reqPayload struct {
		ContactID uint64 `json:"contactId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Prevent a user from adding themselves.
	if reqPayload.ContactID == userID {
		http.Error(w, "cannot add yourself as a contact", http.StatusBadRequest)
		return
	}

	err = rt.db.AddContact(userID, reqPayload.ContactID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// listContacts handles GET /users/:id/contacts.
func (rt *_router) listContacts(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userIdStr := ps.ByName("id")
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	tokenUserID, err := ExtractUserIDFromToken(r)
	if err != nil || tokenUserID != userID {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	contacts, err := rt.db.ListContacts(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(contacts); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode listContacts response")
	}
}

// removeContact handles DELETE /users/:id/contacts/:contactId.
func (rt *_router) removeContact(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userIdStr := ps.ByName("id")
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	tokenUserID, err := ExtractUserIDFromToken(r)
	if err != nil || tokenUserID != userID {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	contactIdStr := ps.ByName("contactId")
	contactID, err := strconv.ParseUint(contactIdStr, 10, 64)
	if err != nil || contactID == 0 {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}
	err = rt.db.RemoveContact(userID, contactID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

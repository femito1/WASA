package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/femito1/WASA/service/api/reqcontext"
	"github.com/femito1/WASA/service/database"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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
	var reqPayload struct {
		CommentText string `json:"commentText"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := rt.db.CheckUserById(database.User{Id: userId})
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	commentId, err := rt.db.CommentMessage(user, convId, msgId, reqPayload.CommentText)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]uint64{"commentId": commentId}); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode comment response")
	}
}

func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userIdStr := ps.ByName("id")
	convIdStr := ps.ByName("convId")
	msgIdStr := ps.ByName("msgId")
	commentIdStr := ps.ByName("commentId")
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
	commentId, err := strconv.ParseUint(commentIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid comment id", http.StatusBadRequest)
		return
	}
	user, err := rt.db.CheckUserById(database.User{Id: userId})
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if err := rt.db.DeleteComment(user, convId, msgId, commentId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

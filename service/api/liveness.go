package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// liveness is an HTTP handler that checks the API server’s status.
// It pings the database and returns 200 if healthy, 500 otherwise.
func (rt *_router) liveness(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := rt.db.Ping(); err != nil {
		rt.baseLogger.WithError(err).Error("liveness check failed: database ping error")
		http.Error(w, "Service Unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write([]byte("OK")); err != nil {
		rt.baseLogger.WithError(err).Error("failed to write liveness response")
	}
}

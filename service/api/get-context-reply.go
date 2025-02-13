package api

import (
	"net/http"

	"github.com/femito1/WASA/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getContextReply is an example of HTTP endpoint that returns "Hello World!" as a plain text. The signature of this
// handler accepts a reqcontext.RequestContext (see httpRouterHandler).
func (rt *_router) getContextReply(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("content-type", "text/plain")
	_, err := w.Write([]byte("Hello World!"))
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to write getContextReply response")
	}
}

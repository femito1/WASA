package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Login
	rt.router.POST("/session", rt.wrap(rt.doLogin))

	// Users
	rt.router.GET("/users", rt.wrap(rt.listUsers))
	rt.router.PUT("/users/:id", rt.wrap(rt.setMyUserName))
	rt.router.PUT("/users/:id/photo", rt.wrap(rt.getContextReply))

	// Conversations
	rt.router.POST("/users/:id/conversations", rt.wrap(rt.getContextReply))
	rt.router.GET("/users/:id/conversations/:convId", rt.wrap(rt.getContextReply))
	rt.router.GET("/users/:id/conversations/:convId/members", rt.wrap(rt.getContextReply))
	rt.router.
		rt.router.
		rt.router.
		rt.router.

		// Messages
		rt.router.
		rt.router.
		rt.router.

		// Comments
		rt.router.
		rt.router.
		rt.router.GET("/context", rt.wrap(rt.getContextReply))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}

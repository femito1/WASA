package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handles the registered APIs.
func (rt *_router) Handler() http.Handler {
	// Login
	rt.router.POST("/session", rt.wrap(rt.doLogin))

	// Users
	rt.router.GET("/users", rt.wrap(rt.listUsers))
	rt.router.PUT("/users/:id", rt.wrap(rt.setMyUserName))
	rt.router.PUT("/users/:id/photo", rt.wrap(rt.setMyPhoto))

	// Conversations
	rt.router.POST("/users/:id/conversations", rt.wrap(rt.createGroup))
	rt.router.GET("/users/:id/conversations", rt.wrap(rt.getMyConversations))
	rt.router.GET("/users/:id/conversations/:convId", rt.wrap(rt.getConversation))
	rt.router.POST("/users/:id/conversations/:convId/members", rt.wrap(rt.addtoGroup))
	rt.router.DELETE("/users/:id/conversations/:convId/members", rt.wrap(rt.leaveGroup))
	rt.router.PUT("/users/:id/conversations/:convId/name", rt.wrap(rt.setGroupName))
	rt.router.PUT("/users/:id/conversations/:convId/photo", rt.wrap(rt.setGroupPhoto))

	// Messages
	rt.router.POST("/users/:id/conversations/:convId/messages", rt.wrap(rt.sendMessage))
	rt.router.DELETE("/users/:id/conversations/:convId/messages/:msgId", rt.wrap(rt.deleteMessage))
	rt.router.POST("/users/:id/conversations/:convId/messages/:msgId/forward", rt.wrap(rt.forwardMessage))

	// Comments
	rt.router.POST("/users/:id/conversations/:convId/messages/:msgId/comment", rt.wrap(rt.commentMessage))
	rt.router.DELETE("/users/:id/conversations/:convId/messages/:msgId/comment/:commentId", rt.wrap(rt.uncommentMessage))
	rt.router.GET("/context", rt.wrap(rt.getContextReply))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}

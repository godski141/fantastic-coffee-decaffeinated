package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	
	// Register routes
	rt.router.GET("/", rt.getHelloWorld)
	rt.router.GET("/context", rt.wrap(rt.getContextReply))


	// Special routes
	rt.router.GET("/liveness", rt.liveness)


	// API routes
	rt.router.POST("/session", rt.doLogin)


	rt.router.POST("/conversations", rt.postConversations)
	rt.router.GET("/conversations", rt.getUserConversations)

	rt.router.GET("/conversations/:convId", rt.getConversationByID)
	rt.router.DELETE("/conversations/:convId", rt.deleteConversation)

	rt.router.POST("/messages", rt.postMessage)
	rt.router.DELETE("/messages/:messageId", rt.deleteMessage)
	rt.router.POST("/messages/:messageId", rt.forwardMessage)
	rt.router.POST("/messages/:messageId/comment", rt.commentMessage)
	rt.router.DELETE("/messages/:messageId/uncomment", rt.unCommentMessage)
	
	return rt.router
}

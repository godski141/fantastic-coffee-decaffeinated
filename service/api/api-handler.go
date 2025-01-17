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
	rt.router.GET("/conversations/start-conversation", rt.getUserConversations)
	rt.router.GET("/conversations/:conversation_id", rt.getConversationByID)
	rt.router.DELETE("/conversations/:conversation_id/delete-conversation", rt.deleteConversation)
	
	rt.router.GET("/conversations/:conv_id/messages", rt.getMessagesFromConversation)
	rt.router.POST("/conversations/:conversation_id/send-message", rt.postMessage)
	rt.router.DELETE("/conversations/:conversation_id/messages/:message_id", rt.deleteMessage)
	rt.router.POST("/conversations/:conversation_id/messages/:message_id", rt.forwardMessage)
	rt.router.POST("/conversations/:conversation_id/messages/:message_id/reaction", rt.commentMessage)
	rt.router.DELETE("/conversations/:conversation_id/messages/:message_id/reaction", rt.unCommentMessage)
	
	rt.router.POST("/conversations/create-group", rt.createGroup)
	
	return rt.router
}

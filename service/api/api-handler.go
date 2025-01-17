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

	rt.router.POST("/conversations/start-conversation", rt.postConversations)
	rt.router.GET("/conversations", rt.getUserConversations)
	rt.router.GET("/conversations/get-details/:conversation_id", rt.getConversationByID)
	rt.router.DELETE("/conversations/delete/:conversation_id", rt.deleteConversation)
	
	rt.router.GET("/conversations/messages/:conversation_id", rt.getMessagesFromConversation)
	rt.router.POST("/conversations/send-message/:conversation_id", rt.postMessage)
	rt.router.DELETE("/conversations/delete-message/:conversation_id/message/:message_id", rt.deleteMessage)
	rt.router.POST("/conversations/forward-message/:conversation_id/messages/:message_id", rt.forwardMessage)
	rt.router.POST("/conversations/react/:conversation_id/messages/:message_id", rt.commentMessage)
	rt.router.DELETE("/conversations/delete-react/:conversation_id/messages/:message_id", rt.unCommentMessage)
	
	rt.router.POST("/conversations/create-group", rt.createGroup)
	
	return rt.router
}

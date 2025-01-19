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
	rt.router.PATCH("/conversations/group/change-name/:conversation_id", rt.renameGroup)
	rt.router.POST("/conversations/group/add/:conversation_id", rt.addToGroup)
	rt.router.DELETE("/conversations/group/leave/:conversation_id", rt.leaveGroup)
	rt.router.PATCH("/conversations/group/change-photo/:conversation_id", rt.updateGroupPhoto)
	rt.router.GET("/conversations/group/get-photo/:conversation_id", rt.getGroupPhoto)

	rt.router.PATCH("/users/modify-username", rt.modifyUserName)
	rt.router.GET("/users/get-photo/:user_id", rt.getUserPhoto)
	rt.router.PATCH("/users/update-photo", rt.updateUserPhoto)
	
	return rt.router
}

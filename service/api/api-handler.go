package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	
	// Register routes
	rt.router.GET("/", rt.getHelloWorld)
	rt.router.GET("/context", rt.wrap(rt.getContextReply, false))


	// Special routes
	rt.router.GET("/liveness", rt.liveness)


	// API routes
	rt.router.POST("/session", rt.wrap(rt.doLogin, false))

	// Conversations
	rt.router.POST("/conversations/start-conversation", rt.wrap(rt.postConversations, true))
	rt.router.GET("/conversations", rt.wrap(rt.getUserConversations, true))
	rt.router.GET("/conversations/get-details/:conversation_id", rt.wrap(rt.getConversationByID, true))
	rt.router.DELETE("/conversations/delete/:conversation_id", rt.wrap(rt.deleteConversation, true))
	rt.router.GET("/conversations/get-members/:conversation_id", rt.wrap(rt.getMembersFromConversation, true))
	
	// Messages
	rt.router.GET("/conversations/messages/:conversation_id", rt.wrap(rt.getMessagesFromConversation, true))
	rt.router.POST("/conversations/send-message/:conversation_id", rt.wrap(rt.postMessage, true))
	rt.router.DELETE("/conversations/delete-message/:conversation_id/message/:message_id", rt.wrap(rt.deleteMessage, true))
	rt.router.POST("/conversations/forward-message/:conversation_id/messages/:message_id", rt.wrap(rt.forwardMessage, true))
	rt.router.POST("/conversations/react/:conversation_id/messages/:message_id", rt.wrap(rt.commentMessage, true))
	rt.router.DELETE("/conversations/delete-react/:conversation_id/messages/:message_id", rt.wrap(rt.unCommentMessage, true))
	
	// Groups
	rt.router.POST("/conversations/create-group", rt.wrap(rt.createGroup, true))
	rt.router.PATCH("/conversations/group/change-name/:conversation_id", rt.wrap(rt.renameGroup, true))
	rt.router.POST("/conversations/group/add/:conversation_id", rt.wrap(rt.addToGroup, true))
	rt.router.DELETE("/conversations/group/leave/:conversation_id", rt.wrap(rt.leaveGroup, true))
	rt.router.PATCH("/conversations/group/change-photo/:conversation_id", rt.wrap(rt.updateGroupPhoto, true))
	rt.router.GET("/conversations/group/get-photo/:conversation_id", rt.wrap(rt.getGroupPhoto, true))
	
	// Users
	rt.router.PATCH("/users/modify-username", rt.wrap(rt.modifyUserName, true))
	rt.router.GET("/users/get-photo/:user_id", rt.wrap(rt.getUserPhoto, true))
	rt.router.PATCH("/users/update-photo", rt.wrap(rt.updateUserPhoto, true))
	
	return rt.router
}

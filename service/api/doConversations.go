package api

import (
	"encoding/json"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"log"
)

// Handler per GET /conversations
func (rt *_router) getUserConversations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID := r.Header.Get("Authorization")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Println("DEBUG: Caricando tutte le conversazioni in una lista per l'utente con ID:", userID)
	conversations, err := rt.db.GetUserConversations(userID)
	if err != nil {
		log.Println("DEBUG: Errore:", err)
		http.Error(w, "Error fetching conversations", http.StatusInternalServerError)
		return
	}
	log.Println("DEBUG: Caricate tutte le conversazioni correttamente per l'utente con ID:", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

// Handler per POST /conversations
func (rt *_router) postConversations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Recupera il creatorId dall'header Authorization
	creatorID := r.Header.Get("Authorization")
	if creatorID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decodifica il requestBody
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := rt.db.GetUserByName(req.Username)

	log.Println("DEBUG: ID:", userID)

	if err!= nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Creazione della conversazione nel database
	convID, convName, convPhoto, err := rt.db.CreatePrivateConversation(creatorID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Photo string `json:"photo"`
	}{
		ID:    convID,
		Name:  convName,
		Photo: convPhoto,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Handler per GET /conversations/{convId}
func (rt *_router) getConversationByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	convID := ps.ByName("convId")

	conversation, err := rt.db.GetConversationByID(convID)
	if err != nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversation)
}

// Handler per DELETE /conversations/{convId}
func (rt *_router) deleteConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	convID := ps.ByName("convId")

	err := rt.db.DeleteConversation(convID)
	if err != nil {
		http.Error(w, "Error deleting conversation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

package api

import (
	"encoding/json"
	"net/http"
	"github.com/julienschmidt/httprouter"
)

type ConvIDResponse struct {
	ID    string `json:"convID"`		
}

type UsernameRequest struct {
	Username string `json:"username"`
}

// Handler per GET /conversations
func (rt *_router) getUserConversations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	// Recupera il creatorId dall'header Authorization
	userID := r.Header.Get("Authorization")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

    // Controlla se l'utente esiste nel database
    _, err := rt.db.GetUserByID(userID)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

	// Recupera le conversazioni dell'utente dal database
	conversations, err := rt.db.GetUserConversations(userID)
	if err != nil {
		http.Error(w, "Error fetching conversations", http.StatusInternalServerError)
		return
	}

	// Invia le conversazioni come risposta
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

// Handler per POST /conversations
func (rt *_router) postConversations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	// Recupera il creatorId dall'header Authorization
	creatorID := r.Header.Get("Authorization")

	// Verifica che il creatorID non sia vuoto
	if creatorID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Controlla se l'utente esiste nel database
	_, err := rt.db.GetUserByID(creatorID)

	// Se l'utente non esiste, ritorna errore
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

	// Decodifica il requestBody
	var req UsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	//Verifica che il campo username non sia vuoto
	if req.Username == "" {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

	// Recupera l'ID dell'utente dal database
	targetUserID, err := rt.db.GetUserByName(req.Username)

	// Se l'utente non esiste, ritorna errore
	if err!= nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Creazione della conversazione nel database
	convID, err := rt.db.CreatePrivateConversation(creatorID, targetUserID)

	// Se la creazione fallisce, ritorna errore
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Invia l'ID della conversazione come risposta
	res := ConvIDResponse{ID : convID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Handler per GET /conversations/{convId}
func (rt *_router) getConversationByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	convID := ps.ByName("convId")

	exists, err := rt.db.ConversationExists(convID)
    if err != nil {
        http.Error(w, "Error checking conversation existence", http.StatusInternalServerError)
        return
    }
    if !exists {
        http.Error(w, "Conversation not found", http.StatusNotFound)
        return
    }

	userID := r.Header.Get("Authorization")

	if userID == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

	isMember, err := rt.db.IsUserInConversation(userID, convID)
    if err != nil {
        http.Error(w, "Error checking conversation membership", http.StatusInternalServerError)
        return
    }

    if !isMember {
        http.Error(w, "Forbidden: You are not a member of this conversation", http.StatusForbidden)
        return
    }

	conversation, err := rt.db.GetConversationByID(convID)
	if err != nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	// Se la conversazione è privata, cambia il nome e la foto in base all'utente
    if conversation.Type == "private" {
        otherUserName, otherUserPhoto, err := rt.db.GetOtherUserDetailsInConversation(convID, userID)
        if err != nil {
            http.Error(w, "Error retrieving other user details", http.StatusInternalServerError)
            return
        }
        conversation.Name = otherUserName
        conversation.Photo = otherUserPhoto
    }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversation)
}

// Handler per DELETE /conversations/{convId}
func (rt *_router) deleteConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    convID := ps.ByName("convId")

    // Recupera l'ID dell'utente autenticato
    userID := r.Header.Get("Authorization")
    if userID == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Controlla se la conversazione esiste
    exists, err := rt.db.ConversationExists(convID)
    if err != nil {
        http.Error(w, "Error checking conversation existence", http.StatusInternalServerError)
        return
    }
    if !exists {
        http.Error(w, "Conversation not found", http.StatusNotFound)
        return
    }

    // Verifica se l'utente è un membro della conversazione
    isMember, err := rt.db.IsUserInConversation(userID, convID)
    if err != nil {
        http.Error(w, "Error checking conversation membership", http.StatusInternalServerError)
        return
    }
    if !isMember {
        http.Error(w, "Forbidden: You are not a member of this conversation", http.StatusForbidden)
        return
    }

    // Procede con l'eliminazione della conversazione
    err = rt.db.DeleteConversation(convID)
    if err != nil {
        http.Error(w, "Error deleting conversation", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"github.com/julienschmidt/httprouter"
)

type MessageRequest struct {
    ConvID string `json:"ConvID"`
    Text string `json:"content"`
}

type ConversationsRequest struct {
    ID string `json:"Id"`
}

// postMessage handles POST /messages
func (rt *_router) postMessage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

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

    // Decodifica il body della richiesta
    var req MessageRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

    // Controlla che i campi ConvID e Text non siano vuoti
    if req.ConvID == "" {
        http.Error(w, "ConvID cannot be empty", http.StatusBadRequest)
        return
    }

    // Verifica che la conversazione esista
    exist, err := rt.db.ConversationExists(req.ConvID)
    if err != nil {
        http.Error(w, "Error checking conversation existence", http.StatusInternalServerError)
        return
    }
    if !exist {
        http.Error(w, "Conversation not found", http.StatusNotFound)
        return
    }

    // Verifica se l'utente è un membro della conversazione
    isMember, err := rt.db.IsUserInConversation(userID, req.ConvID)
    if err != nil {
        http.Error(w, "Error checking conversation membership", http.StatusInternalServerError)
        return
    }
    if !isMember {
        http.Error(w, "Forbidden: You are not a member of this conversation", http.StatusForbidden)
        return
    }

    if req.Text == "" {
        http.Error(w, "Text cannot be empty", http.StatusBadRequest)
        return
    }

    // Inserisce il messaggio nel database
    messageID, err := rt.db.InsertMessage(req.ConvID, userID, req.Text)

    if err != nil {
        http.Error(w, "Error inserting message", http.StatusInternalServerError)
        return
    }

    // Aggiorna l'ultimo messaggio della conversazione
    if err := rt.db.UpdateLastMessage(req.ConvID, messageID); err != nil {
        log.Println(err)
        http.Error(w, "Error updating last message", http.StatusInternalServerError)
        return
    }

    // Recupera il messaggio completo dal database
    messageResponse, err := rt.db.GetMessageFromID(messageID)

    if err != nil {
        http.Error(w, "Error fetching message", http.StatusInternalServerError)
        return
    }

    // Invia il messaggio come risposta
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(messageResponse)
}

// deleteMessage handles DELETE /messages/:messageID
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

    // Recupera l'ID del messaggio dalla richiesta
    messageID := ps.ByName("messageId")

    // Verifica che l'utente sia il mittente del messaggio
    message, err := rt.db.GetMessageFromID(messageID)
    log.Println("DEBUG: message->", message)
    log.Println("DEBUG: message.SenderID->", message.SenderID)
    log.Println("DEBUG: err->", err)
    if err != nil {
        // Se il messaggio non esiste restituisce 404
        if errors.Is(err, sql.ErrNoRows) {
        http.Error(w, "Message not found", http.StatusNotFound)
    } else {

        // Altrimenti restituisce un errore interno del server
        http.Error(w, "Error fetching message", http.StatusInternalServerError)
    }
        return
    }
    if message.SenderID != userID {
        http.Error(w, "Forbidden: You are not the sender of this message", http.StatusForbidden)
        return
    }

    // Cancella il messaggio dal database
    if err := rt.db.DeleteMessage(messageID); err != nil {
        http.Error(w, "Error deleting message", http.StatusInternalServerError)
        return
    }

    // Trova il nuovo ultimo messaggio della conversazione
    newLastMessageID, err := rt.db.GetLastMessageID(message.ConversationID)
    if err != nil {
        log.Println("Error retrieving last message:", err)
        http.Error(w, "Error updating last message", http.StatusInternalServerError)
        return
    }

    // Aggiorna il lastMessageId della conversazione
    if err := rt.db.UpdateLastMessage(message.ConversationID, newLastMessageID); err != nil {
        log.Println("Error updating conversation last message:", err)
        http.Error(w, "Error updating last message", http.StatusInternalServerError)
        return
    }

    // Invia una risposta vuota
    w.WriteHeader(http.StatusNoContent)
}

// postMessage handles POST /messages/:messageID
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    
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

    // Recupera l'ID del messaggio dalla richiesta
    messageID := ps.ByName("messageId")

    // Verifica che il messaggio esista
    message, err := rt.db.GetMessageFromID(messageID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            http.Error(w, "Message not found", http.StatusNotFound)
        } else {
            http.Error(w, "Error fetching message", http.StatusInternalServerError)
        }
        return
    }

    // Verifica che l'utente sia un membro della conversazione
    isMember, err := rt.db.IsUserInConversation(userID, message.ConversationID)
    if err != nil {
        http.Error(w, "Error checking conversation membership", http.StatusInternalServerError)
        return
    }
    if !isMember {
        http.Error(w, "Forbidden: You are not a member of this conversation", http.StatusForbidden)
        return
    }

    // Lettura del body della richiesta
    var req ConversationsRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Verifica che il campo Id non sia vuoto
    if req.ID == "" {
        http.Error(w, "Destination conversation ID cannot be empty", http.StatusBadRequest)
        return
    }

    // Verifica che la conversazione esista
    exist, err := rt.db.ConversationExists(req.ID)
    if err != nil {
        http.Error(w, "Error checking conversation existence", http.StatusInternalServerError)
        return
    }
    if !exist {
        http.Error(w, "Conversation not found", http.StatusNotFound)
        return
    }

    // Verifica che l'utente sia un membro della conversazione
    isMember, err = rt.db.IsUserInConversation(userID, req.ID)
    if err != nil {
        http.Error(w, "Error checking conversation membership", http.StatusInternalServerError)
        return
    }
    if !isMember {
        http.Error(w, "Forbidden: You are not a member of this conversation", http.StatusForbidden)
        return
    }

    // Inserisce il messaggio nel database
    newMessageID, err := rt.db.InsertMessage(req.ID, userID, message.Content)
    if err != nil {
        http.Error(w, "Error inserting message", http.StatusInternalServerError)
        return
    }

    // Aggiorna l'ultimo messaggio della conversazione
    if err := rt.db.UpdateLastMessage(req.ID, newMessageID); err != nil {
        log.Println(err)
        http.Error(w, "Error updating last message", http.StatusInternalServerError)
        return
    }

    // Recupera il messaggio completo dal database
    messageResponse, err := rt.db.GetMessageFromID(newMessageID)
    if err != nil {
        http.Error(w, "Error fetching message", http.StatusInternalServerError)
        return
    }

    // Invia il messaggio come risposta
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(messageResponse)
}



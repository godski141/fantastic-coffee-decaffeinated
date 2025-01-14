package api

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/julienschmidt/httprouter"
)

type MessageRequest struct {
    ConvID string `json:"ConvID"`
    Text string `json:"content"`
}


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


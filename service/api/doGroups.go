package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type GroupRequest struct {
	Name string `json:"name"`
	Members []string `json:"members"`
}

// createGroup handles POST /groups/create-group
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	
	// Recupera l'userId dal Authorization Header
	userID := r.Header.Get("Authorization")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Controllo se l'utente esiste nel database
	_, err := rt.db.GetUserByID(userID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decodifica il body della richiesta
	var req GroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Controllo se il nome del gruppo è vuoto
	if req.Name == "" {
		http.Error(w, "Invalid group name", http.StatusBadRequest)
		return
	}

	// Controllo se il gruppo ha almeno un membro
	if len(req.Members) == 0 {
		http.Error(w, "Group must have at least one member", http.StatusBadRequest)
		return
	}

	// Controllo se gli utenti esistono nel database
	var invalidMembers []string
	for _, memberName := range req.Members {
		
		_, err := rt.db.GetUserByName(strings.ToLower(memberName))
		if err != nil {
			log.Println(err)
			invalidMembers = append(invalidMembers, strings.ToLower(memberName))
		}
	}
	if len(invalidMembers) > 0 {
		http.Error(w, "Invalid members: "+strings.Join(invalidMembers, ", "), http.StatusBadRequest)
		return
	}

	// Creazione del gruppo
	groupId, err := rt.db.CreateGroup(req.Name, userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Aggiungo il creatore al gruppo
	if err := rt.db.AddUserToGroup(groupId, userID); err != nil {
		log.Println(err)
		rt.db.DeleteConversation(groupId)
		http.Error(w, "Error adding creator to group", http.StatusInternalServerError)
		return
	}

	// Aggiungo i membri al gruppo
	for _, memberName := range req.Members {
		memberId, _ := rt.db.GetUserByName(strings.ToLower(memberName))
		if err := rt.db.AddUserToGroup(groupId, memberId); err != nil {
			rt.db.DeleteConversation(groupId)
			http.Error(w, "Error adding members to group", http.StatusInternalServerError)
			return
		}
	}
	
	res := ConvIDResponse{ConversationID: groupId}

	// Risposta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}
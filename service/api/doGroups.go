package api

import (
	"WasaTEXT/service/api/reqcontext"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/julienschmidt/httprouter"
)

type GroupRequest struct {
	Name string `json:"name"`
	Members []string `json:"members"`
}

type NewGroupName struct {
	Name string `json:"name"`
}

// createGroup handles POST /groups/create-group
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	
	// Recupera l'userId dal Authorization Header
	userID := ctx.UserId

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

// renameGroup handles PATCH conversations/groups/change-name/:groupId
func (rt *_router) renameGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Recupera l'userId dal Authorization Header
	userID := ctx.UserId

	// Recupera il groupId dai parametri
	groupID := ps.ByName("conversation_id")

	// Controllo se il gruppo esiste
	exist , err := rt.db.ConversationExists(groupID)
	if err != nil {
		http.Error(w, "Error", http.StatusNotFound)
		return
	}
	if !exist {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Se esiste, controllo se è un gruppo
	isPrivate, err := rt.db.IsConversationPrivate(groupID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if isPrivate {
		http.Error(w, "Conversation is not a group", http.StatusBadRequest)
		return
	}

	// Controllo se l'utente è il creatore del gruppo
	isCreator, err := rt.db.IsUserCreatorOfGroup(userID, groupID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if !isCreator {
		http.Error(w, "Forbidden: You are not the creator of this group", http.StatusForbidden)
		return
	}


	// Decodifica il body della richiesta
	var req NewGroupName
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Controllo se il nome del gruppo è vuoto
	if req.Name == "" {
		http.Error(w, "Invalid group name", http.StatusBadRequest)
		return
	}

	// Controllo se il nome del gruppo è valido
	if len(req.Name) < 3 || len(req.Name) > 50 {
		http.Error(w, "Group name must be between 3 and 50 characters", http.StatusBadRequest)
		return
	}

	// Cambio il nome del gruppo
	err = rt.db.ChangeGroupName(groupID, req.Name)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Risposta
	w.WriteHeader(http.StatusNoContent)
}

// addToGroup handles POST conversations/groups/add-user/:groupId
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Recupera l'userId dal Authorization Header
	userID := ctx.UserId

	// Recupera il groupId dai parametri
	groupID := ps.ByName("conversation_id")

	// Controllo se il gruppo esiste
	exist , err := rt.db.ConversationExists(groupID)
	if err != nil {
		http.Error(w, "Error", http.StatusNotFound)
		return
	}
	if !exist {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Se esiste, controllo se è un gruppo
	isPrivate, err := rt.db.IsConversationPrivate(groupID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if isPrivate {
		http.Error(w, "Conversation is not a group", http.StatusBadRequest)
		return
	}

	// Controllo se l'utente è nel gruppo
	isMember, err := rt.db.IsUserInConversation(userID, groupID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Who is trying to add, is not in the selected group", http.StatusBadRequest)
		return
	}

	// Decodifica il body della richiesta
	var req struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	lowerUsername := strings.ToLower(req.Username)
	// Controllo se l'utente esiste nel database
	user2ID , err := rt.db.GetUserByName(lowerUsername)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Controllo se l'utente è già nel gruppo
	isMember, err = rt.db.IsUserInConversation(user2ID, groupID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if isMember {
		http.Error(w, "User is already in the group", http.StatusBadRequest)
		return
	}

	// Aggiungo l'utente al gruppo
	err = rt.db.AddUserToGroup(groupID, user2ID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Risposta
	w.WriteHeader(http.StatusNoContent)
}

// leaveGroup handles DELETE conversations/groups/leave/:groupId
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Recupera l'userId dal Authorization Header
	userID := ctx.UserId

	// Recupera il groupId dai parametri
	groupID := ps.ByName("conversation_id")

	// Controllo se il gruppo esiste
	exist , err := rt.db.ConversationExists(groupID)
	if err != nil {
		http.Error(w, "Error", http.StatusNotFound)
		return
	}
	if !exist {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Se esiste, controllo se è un gruppo
	isPrivate, err := rt.db.IsConversationPrivate(groupID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if isPrivate {
		http.Error(w, "Conversation is not a group", http.StatusBadRequest)
		return
	}

	// Controllo se l'utente è nel gruppo
	isMember, err := rt.db.IsUserInConversation(userID, groupID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "User is not in the group", http.StatusBadRequest)
		return
	}

	// Controllo se l'utente è il creatore del gruppo
	isCreator, err := rt.db.IsUserCreatorOfGroup(userID, groupID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if isCreator {
		http.Error(w, "Forbidden: Creator cannot leave the group", http.StatusForbidden)
		return
	}

	// Tolgo l'utente dal gruppo
	err = rt.db.LeaveGroup(groupID, userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Risposta
	w.WriteHeader(http.StatusNoContent)
}

// updatePhotoGroup handles PATCH conversations/groups/update-photo/:groupId 
func (rt *_router) updateGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
    // Recupera l'userId dal Authorization Header
    userID := ctx.UserId

    // Recupera il groupId dai parametri
    groupID := ps.ByName("conversation_id")

    // Controlla se il gruppo esiste
    exists, err := rt.db.ConversationExists(groupID)
    if err != nil || !exists {
        http.Error(w, "Group not found", http.StatusNotFound)
        return
    }

	// Controlla se il gruppo è di tipo "group"
	isPrivate, err := rt.db.IsConversationPrivate(groupID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if isPrivate {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

    // Controlla se l'utente fa parte del gruppo
	isMember, err := rt.db.IsUserInConversation(userID, groupID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Forbidden: You are not a member of this group", http.StatusForbidden)
		return
	}

    // Decodifica il body della richiesta
    var req struct {
        PhotoBase64 string `json:"photo"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Verifica che la foto non sia vuota
    if req.PhotoBase64 == "" {
        http.Error(w, "Photo cannot be empty", http.StatusBadRequest)
        return
    }

	// Rimuovi l'intestazione "data:image/png;base64," o "data:image/jpeg;base64,"
    photoData := strings.TrimPrefix(req.PhotoBase64, "data:image/png;base64,")
    photoData = strings.TrimPrefix(photoData, "data:image/jpeg;base64,")

    // Decodifica l'immagine Base64
    decodedPhoto, err := base64.StdEncoding.DecodeString(photoData)
    if err != nil {
        http.Error(w, "Invalid Base64 string", http.StatusBadRequest)
        return
    }

    // Percorso per la directory e il file
	dirPath := "service/uploads/groups/"
	filePath := fmt.Sprintf("%s%s_photo.png", dirPath, groupID)

	// Assicurati che la directory esista
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		log.Println("Failed to create directory:", err)
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	// Salva l'immagine nel file
	err = os.WriteFile(filePath, decodedPhoto, 0644)
	if err != nil {
		log.Println("Failed to save image:", err)
		http.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}


    // Aggiorna il percorso della foto nel database
    err = rt.db.UpdateGroupPhoto(groupID, filePath)
    if err != nil {
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }

    // Rispondi con successo
    w.WriteHeader(http.StatusNoContent)
}

// getGroupPhoto handles GET conversations/groups/photo/:conversation_id
func (rt *_router) getGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Recupera l'userId dal Authorization Header
	userID := ctx.UserId

    // Recupera l'ID della conversazione
    conversationID := ps.ByName("conversation_id")


    // Verifica che la conversazione esista e sia di tipo "group"
	isPrivate, err := rt.db.IsConversationPrivate(conversationID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if isPrivate {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Verifica che l'utente faccia parte del gruppo
	isMember, err := rt.db.IsUserInConversation(userID, conversationID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Forbidden: You are not a member of this group", http.StatusForbidden)
		return
	}

	// Recupera il percorso della foto dal database
	photoPath, err := rt.db.GetGroupPhotoByID(conversationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Group not found or photo not set", http.StatusNotFound)
			return
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

    // Usa una foto predefinita se il campo photo è NULL o vuoto
	if photoPath == "" {
		photoPath = "service/uploads/default_user_photo.jpg"
	}

    // Serve il file immagine
    http.ServeFile(w, r, photoPath)
}

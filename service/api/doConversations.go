package api

import (
    "WasaTEXT/service/api/reqcontext"
    "encoding/json"
    "log"
    "net/http"
    "strings"

    "github.com/julienschmidt/httprouter"
)

type ConvIDResponse struct {
    ConversationID string `json:"conversation_id"`
}

type UsernameRequest struct {
    Username string `json:"username"`
}

// Handler per GET /conversations
func (rt *_router) getUserConversations(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {

    // Recupera l'ID dell'utente autenticato
    userID := ctx.UserId

    // Recupera le conversazioni dell'utente dal database
    conversations, err := rt.db.GetUserConversations(userID)
    if err != nil {
        log.Println(err)
        http.Error(w, "Error fetching conversations", http.StatusInternalServerError)
        return
    }

    // Invia le conversazioni come risposta
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(conversations)
}

// Handler per POST /conversations/start-conversation
func (rt *_router) postConversations(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {

    // Recupera l'ID dell'utente autenticato
    creatorID := ctx.UserId

    // Decodifica il requestBody
    var req UsernameRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        ctx.Logger.Error("Invalid request body: ", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Verifica che il campo username non sia vuoto
    if req.Username == "" {
        ctx.Logger.Error("Invalid request body: username is empty")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    log.Println("DEBUG: Nome username target: ", req.Username)

    // Converte il nome dell'utente in minuscolo
    lowername := strings.ToLower(req.Username)

    // Recupera l'ID dell'utente dal database
    targetUserID, err := rt.db.GetUserByName(lowername)

    // Se l'utente non esiste, ritorna errore
    if err != nil {
        ctx.Logger.Error("User not found: ", err)
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    // Creazione della conversazione nel database
    convID, err := rt.db.CreatePrivateConversation(creatorID, targetUserID)

    log.Println("DEBUG: ID della conversazione appena creata:", convID)

    // Se la creazione fallisce, ritorna errore
    if err != nil {
        ctx.Logger.Error("Error creating conversation: ", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Invia l'ID della conversazione come risposta
    res := ConvIDResponse{ConversationID: convID}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(res)
}

// Handler per GET /conversations/{convId}
func (rt *_router) getConversationByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

    userID := ctx.UserId

    // Recupera l'ID della conversazione
    convID := ps.ByName("conversation_id")

    exists, err := rt.db.ConversationExists(convID)
    if err != nil {
        http.Error(w, "Error checking conversation existence", http.StatusInternalServerError)
        return
    }
    if !exists {
        http.Error(w, "Conversation not found", http.StatusNotFound)
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

    conversation, err := rt.db.GetConversationByID(convID, userID)
    log.Println("DEBUG: error:", err)
    if err != nil {
        http.Error(w, "Conversation not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(conversation)
}

// Handler per DELETE /conversations/{convId}/delete
func (rt *_router) deleteConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
    convID := ps.ByName("conversation_id")

    // Recupera l'ID dell'utente autenticato
    userID := ctx.UserId

    // Controlla se la conversazione esiste
    exists, err := rt.db.ConversationExists(convID)
    if err != nil {
        log.Println("Error checking conversation existence:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    if !exists {
        http.Error(w, "Conversation not found", http.StatusNotFound)
        return
    }

    // Verifica se la conversazione è privata o di gruppo
    isPrivate, err := rt.db.IsConversationPrivate(convID)
    if err != nil {
        log.Println("Error checking conversation type:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    if isPrivate {
        // Controlla che l'utente sia un membro
        isMember, err := rt.db.IsUserInConversation(userID, convID)
        if err != nil {
            log.Println("Error checking membership:", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
        if !isMember {
            http.Error(w, "Forbidden: You are not a member of this conversation", http.StatusForbidden)
            return
        }
    } else {
        // Controlla che l'utente sia il creatore
        isCreator, err := rt.db.IsUserCreatorOfGroup(userID, convID)
        if err != nil {
            log.Println("Error checking conversation creator:", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
        if !isCreator {
            http.Error(w, "Forbidden: You are not the creator of this conversation", http.StatusForbidden)
            return
        }
    }

    // Elimina la conversazione
    err = rt.db.DeleteConversation(convID)
    if err != nil {
        log.Println("Error deleting conversation:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    log.Println("Conversation deleted successfully:", convID)
    w.WriteHeader(http.StatusNoContent)
}

// Handler per GET /conversations/get-members/:conversion_id
func (rt *_router) getMembersFromConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
    // Recupera l'ID dell'utente autenticato
    userID := ctx.UserId

    // Recupera l'ID della conversazione
    convID := ps.ByName("conversation_id")

    // Controlla se la conversazione esiste
    exists, err := rt.db.ConversationExists(convID)
    if err != nil {
        ctx.Logger.WithError(err).Error("Error checking conversation existence")
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    if !exists {
        http.Error(w, "Conversation not found", http.StatusNotFound)
        return
    }

    // Controlla che l'utente sia un membro
    isMember, err := rt.db.IsUserInConversation(userID, convID)
    if err != nil {
        ctx.Logger.WithError(err).Error("Error checking conversation membership")
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    if !isMember {
        http.Error(w, "Forbidden: You are not a member of this conversation", http.StatusForbidden)
        return
    }

    // Recupera i membri della conversazione
    members, err := rt.db.GetMembersFromConversation(convID)
    if err != nil {
        ctx.Logger.WithError(err).Error("Error fetching members")
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Invia i membri come risposta
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(members)

}
package api

import (
	"WasaTEXT/service/api/reqcontext"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"unicode/utf8"

	"github.com/julienschmidt/httprouter"
)

type MessageRequest struct {
    Text string `json:"content"`
}

type ConversationsRequest struct {
    ID string `json:"conversation_id"`
}

type ReactionRequest struct {
    Reaction string `json:"emoji"`
}

// postMessage handles POST /conversations/:conversation_id/send-message
func (rt *_router) postMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

    // Recupera l'userID dall'header Authorization
    userID := ctx.UserId

    // Recupera l'ID della conversazione dal parametro URL
    convID := ps.ByName("conversation_id")
    if convID == "" {
        http.Error(w, "Conversation ID cannot be empty", http.StatusBadRequest)
        return
    }

    // Verifica che la conversazione esista
    exist, err := rt.db.ConversationExists(convID)
    if err != nil {
        http.Error(w, "Error checking conversation existence", http.StatusInternalServerError)
        return
    }
    if !exist {
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

    // Decodifica il body della richiesta
    var req struct {
        Text string `json:"content"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Controlla che il testo del messaggio non sia vuoto
    if req.Text == "" {
        http.Error(w, "Message content cannot be empty", http.StatusBadRequest)
        return
    }

    // Inserisce il messaggio nel database
    messageID, err := rt.db.InsertMessage(convID, userID, req.Text)
    if err != nil {
        http.Error(w, "Error inserting message", http.StatusInternalServerError)
        return
    }

    // Aggiorna l'ultimo messaggio della conversazione
    if err := rt.db.UpdateLastMessage(convID, messageID); err != nil {
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
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(messageResponse)
}

// deleteMessage handles DELETE /conversations/:conversation_id/messages/:message_id
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
    // Recupera il creatorId dall'header Authorization
	userID := ctx.UserId


    // Recupera l'ID del messaggio dalla richiesta
    messageID := ps.ByName("message_id")

    // Recupera l'ID della conversazione dalla richiesta
    convID := ps.ByName("conversation_id")

    // Verifica che la conversazione esista
    exist, err := rt.db.ConversationExists(convID)
    if err != nil {
        http.Error(w, "Error checking conversation existence", http.StatusInternalServerError)
        return
    }
    if !exist {
        http.Error(w, "Conversation not found", http.StatusNotFound)
        return
    }

    // Verifica che l'utente appartenga alla conversazione 
    isMember, err := rt.db.IsUserInConversation(userID, convID)
    if err != nil {
        http.Error(w, "Error checking conversation membership", http.StatusInternalServerError)
        return
    }
    if !isMember {
        http.Error(w, "Forbidden: You are not a member of this conversation", http.StatusForbidden)
        return
    }

    // Verifica che il messaggio esista
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

    // Verifica che il messaggio appartenga alla conversazione
    if message.ConversationID != convID {
        http.Error(w, "Forbidden: Message does not belong to this conversation", http.StatusForbidden)
        return
    }

    // Verifica che l'utente sia il mittente del messaggio
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

// postMessage handles POST /conversations/:conversation_id/messages/:message_id
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
    
    // Recupera il creatorId dall'header Authorization
    userID := ctx.UserId

    // Recupera l'ID del messaggio dalla richiesta
    messageID := ps.ByName("message_id")

    // Recupera l'ID della conversazione dalla richiesta
    convID := ps.ByName("conversation_id")

    // Verifica che la conversazione esista
    exist, err := rt.db.ConversationExists(convID)
    if err != nil {
        http.Error(w, "Error checking conversation existence", http.StatusInternalServerError)
        return
    }
    if !exist {
        http.Error(w, "Conversation not found", http.StatusNotFound)
        return
    }

    // Verifica che l'utente sia un membro della conversazione
    isMember, err := rt.db.IsUserInConversation(userID, convID)
    if err != nil {
        http.Error(w, "Error checking conversation membership", http.StatusInternalServerError)
        return
    }
    if !isMember {
        http.Error(w, "Forbidden: You are not a member of this conversation", http.StatusForbidden)
        return
    }

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

    // Verifica che il messaggio appartenga alla conversazione
    if message.ConversationID != convID {
        http.Error(w, "Forbidden: Message does not belong to this conversation", http.StatusForbidden)
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
    exist, err = rt.db.ConversationExists(req.ID)
    if err != nil {
        http.Error(w, "Error checking conversation existence", http.StatusInternalServerError)
        return
    }
    if !exist {
        http.Error(w, "Conversation not found", http.StatusNotFound)
        return
    }

    // Verifica che l'utente sia un membro della conversazione di destinazione
    isMember, err = rt.db.IsUserInConversation(userID, req.ID)
    if err != nil {
        http.Error(w, "Error checking conversation membership", http.StatusInternalServerError)
        return
    }
    if !isMember {
        http.Error(w, "Forbidden: You are not a member of the target conversation", http.StatusForbidden)
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

// commentMessage handles POST /conversations/:conversation_id/messages/:message_id/reaction
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
    // Recupera il creatorId dall'header Authorization
    userID := ctx.UserId

    // Recupera l'ID del messaggio dalla richiesta
    messageID := ps.ByName("message_id")

    // Recupera l'ID della conversazione dalla richiesta
    convID := ps.ByName("conversation_id")

    // Verifica che la conversazione esista
    exist, err := rt.db.ConversationExists(convID)
    if err != nil {
        http.Error(w, "Error checking conversation existence", http.StatusInternalServerError)
        return
    }
    if !exist {
        http.Error(w, "Conversation not found", http.StatusNotFound)
        return
    }

    // Verifica che l'utente sia un membro della conversazione associata al messaggio
    isMember, err := rt.db.IsUserInConversation(userID, convID)
    if err != nil {
        http.Error(w, "Error checking conversation membership", http.StatusInternalServerError)
        return
    }
    if !isMember {
        http.Error(w, "Forbidden: You are not a member of the message's conversation", http.StatusForbidden)
        return
    }

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

    // Verifica che il messaggio appartenga alla conversazione
    if message.ConversationID != convID {
        http.Error(w, "Forbidden: Message does not belong to this conversation", http.StatusForbidden)
        return
    }

    // Lettura del body della richiesta
    var req ReactionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Verifica che il campo reaction sia un emoji
    if !isSingleEmoji(req.Reaction) {
        http.Error(w, "Invalid reaction: must be an emoji and not a combination of emojis", http.StatusBadRequest)
    }
    

    // Inserisci la reazione nel database associando a userID e messageID
    if err := rt.db.InsertReaction(userID, messageID, req.Reaction); err != nil {
        http.Error(w, "Error inserting reaction", http.StatusInternalServerError)
        return
    }

    // Invia una risposta vuota
    w.WriteHeader(http.StatusNoContent)
}

// unCommentMessage handles DELETE /conversations/:conversation_id/messages/:message_id/reaction
func (rt *_router) unCommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
    // Recupera l'userID dall'header Authorization
    userID := ctx.UserId

    // Recupera l'ID del messaggio dalla richiesta
    messageID := ps.ByName("message_id")

    // Recupera l'ID della conversazione dalla richiesta
    convID := ps.ByName("conversation_id")

    // Verifica che la conversazione esista
    exist, err := rt.db.ConversationExists(convID)
    if err != nil {
        http.Error(w, "Error checking conversation existence", http.StatusInternalServerError)
        return
    }
    if !exist {
        http.Error(w, "Conversation not found", http.StatusNotFound)
        return
    }

    // Verifica che l'utente sia un membro della conversazione associata al messaggio
    isMember, err := rt.db.IsUserInConversation(userID, convID)
    if err != nil {
        http.Error(w, "Error checking conversation membership", http.StatusInternalServerError)
        return
    }
    if !isMember {
        http.Error(w, "Forbidden: You are not a member of the message's conversation", http.StatusForbidden)
        return
    }

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

    // Verifica che il messaggio appartenga alla conversazione
    if message.ConversationID != convID {
        http.Error(w, "Forbidden: Message does not belong to this conversation", http.StatusForbidden)
        return
    }

    // Controlla se l'utente ha già reagito al messaggio
    hasReaction, err := rt.db.UserHasReaction(messageID, userID)
    if err != nil {
        http.Error(w, "Error checking reaction", http.StatusInternalServerError)
        return
    }
    if !hasReaction {
        http.Error(w, "Reaction not found", http.StatusNotFound)
        return
    }

    // Elimina la reazione dal database
    if err := rt.db.DeleteReaction(messageID, userID); err != nil {
        http.Error(w, "Error deleting reaction", http.StatusInternalServerError)
        return
    }

    // Invia una risposta vuota
    w.WriteHeader(http.StatusNoContent)
}

// Handler per GET /conversations/{convId}/messages
func (rt *_router) getMessagesFromConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
    // Recupera l'userID dall'header Authorization
    userID := ctx.UserId

    // Recupera l'ID della conversazione dalla richiesta
    conversationID := ps.ByName("conversation_id")

    // Verifica che la conversazione esista
    exists, err := rt.db.ConversationExists(conversationID)
    if err != nil {
        http.Error(w, "Error checking conversation existence", http.StatusInternalServerError)
        return
    }
    if !exists {
        http.Error(w, "Conversation not found", http.StatusNotFound)
        return
    }

    // Verifica che l'utente sia un membro della conversazione
    isMember, err := rt.db.IsUserInConversation(userID, conversationID)
    if err != nil {
        http.Error(w, "Error checking conversation membership", http.StatusInternalServerError)
        return
    }
    if !isMember {
        http.Error(w, "Forbidden: You are not a member of this conversation", http.StatusForbidden)
        return
    }

    // Recupera tutti i messaggi della conversazione
    messages, err := rt.db.GetMessagesFromConversation(conversationID)
    if err != nil {
        http.Error(w, "Error fetching messages", http.StatusInternalServerError)
        return
    }

    // Invia i messaggi come risposta
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(messages)
}

// isSingleEmoji verifica se una stringa è un singolo emoji
func isSingleEmoji(input string) bool {
	// Decodifica il primo carattere Unicode
	r, size := utf8.DecodeRuneInString(input)

	// Verifica che il carattere sia valido e che la lunghezza totale della stringa corrisponda
	return r != utf8.RuneError && size == len(input) && isEmojiRune(r)
}

// isEmojiRune verifica se un carattere è una emoji valida
func isEmojiRune(r rune) bool {
	// Intervalli Unicode per le emoji
	return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticon
		(r >= 0x1F300 && r <= 0x1F5FF) || // Simboli e pittogrammi
		(r >= 0x1F680 && r <= 0x1F6FF) || // Trasporti e simboli
		(r >= 0x2600 && r <= 0x26FF) ||   // Simboli vari
		(r >= 0x2700 && r <= 0x27BF)      // Dingbats
}

package database

import (
	"database/sql"
	"fmt"
)

// GetUserConversations recupera tutte le conversazioni di un utente
func (db *appdbimpl) GetUserConversations(userID string) ([]Conversation, error) {
    // Esegui la query SQL per recuperare gli id delle conversazioni dell'utente
    rows, err := db.c.Query(`
        SELECT DISTINCT c.id 
        FROM conversations c
        LEFT JOIN group_members gm ON c.id = gm.conversation_id
        WHERE (c.creator_id = ? OR c.otherUser = ? AND c.type = 'private') OR gm.user_id = ?`, userID, userID, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Inizializza una slice di conversazioni vuota
    var conversations []Conversation
    // Per ogni conversazione trovata
    for rows.Next() {
        var convID string
        // Leggi l'ID della conversazione
        if err := rows.Scan(&convID); err != nil {
            return nil, err
        }
        // Recupera i dettagli della conversazione
        conv, err := db.GetConversationByID(convID, userID)
        if err != nil {
            return nil, err
        }
        // Aggiungi la conversazione alla slice
        conversations = append(conversations, conv)
    }

    // Restituisci la slice di conversazioni
    return conversations, nil
}

// GetConversationByID retrieves a specific conversation from the database by its ID.
func (db *appdbimpl) GetConversationByID(convID, userID string) (Conversation, error) {
    var conv Conversation
    var lastMessageID sql.NullString
    var name sql.NullString
    var photo sql.NullString
    var otherUser sql.NullString

    // Esegui la query SQL per recuperare la conversazione
    err := db.c.QueryRow(`
    SELECT  c.name, c.type, c.creator_id, c.photo, c.lastMessageId, c.otherUser
    FROM conversations c
    WHERE c.id = ?`, convID).Scan(&name, &conv.Type, &conv.CreatorID, &photo, &lastMessageID, &otherUser)

    // Restituisci un errore se la query fallisce
    if err != nil {
        return conv, err
    }

    // Controlla se la conversazione è privata
    if conv.Type == "private" {
        // Controlla se l'utente è il creatore o l'altro utente della conversazione
        if userID == conv.CreatorID {
            otherUserName, err := db.GetUserByID(otherUser.String)
            if err != nil {
                return conv, err
            }
            conv.Name = otherUserName
            conv.Photo = fmt.Sprintf("/users/get-photo/%s", otherUser.String) // Endpoint foto utente
        } else {
            creatorName, err := db.GetUserByID(conv.CreatorID)
            if err != nil {
                return conv, err
            }
            conv.Name = creatorName
            conv.Photo = fmt.Sprintf("/users/get-photo/%s", conv.CreatorID) // Endpoint foto utente
        }
    } else {
        // Se la conversazione è di gruppo, usa il nome e l'endpoint della foto della conversazione
        conv.Name = name.String
        conv.Photo = fmt.Sprintf("/conversations/group/get-photo/%s", convID)
    }

    // Recupera l'ultimo messaggio della conversazione
    if lastMessageID.String != "" {
        msg, err := db.GetContentFromMessageID(lastMessageID.String)
        if err != nil {
            return conv, err
        }
        conv.LastMessage = msg
    } else {
        conv.LastMessage = ""
    }

    conv.ConvID = convID

    // Restituisci la conversazione
    return conv, nil
}

// DeleteConversation deletes a conversation from the database by its ID.
func (db *appdbimpl) DeleteConversation(convID string) error {
    // Execute the SQL command to delete the conversation by ID
    _, err := db.c.Exec(`DELETE FROM conversations WHERE id = ?`, convID)
    if err != nil {
        return err 
    }

    return nil 
}

// CreatePrivateConversation crea una nuova conversazione privata tra due utenti
func (db *appdbimpl) CreatePrivateConversation(user1 string, user2 string) (string, error) {

    // Controllo se l'utente sta cercando di creare una conversazione con se stesso
    if user1 == user2 {
        return "", fmt.Errorf("400: cannot create a conversation with yourself")
    }

    // Controlla se la conversazione esiste già con user1 come creator_id e user2 come other_user (o viceversa)
    var convID string
    err := db.c.QueryRow(`
        SELECT id 
        FROM conversations
        WHERE 
            (creator_id = ? AND otherUser = ?) OR 
            (creator_id = ? AND otherUser = ?) AND type = 'private'
    `, user1, user2, user2, user1).Scan(&convID)

    if err==nil{
        return convID, nil
    }

    // Inizia una transazione per garantire la coerenza
    tx, err := db.c.Begin()
    if err != nil {
        return "", err
    }

    // Inserisce una nuova conversazione nella tabella conversations
    result, err := tx.Exec(`
        INSERT INTO conversations (name, type, creator_id, photo, lastMessageId, otherUser)
        VALUES (NULL, 'private', ?, NULL, NULL, ?)`, user1, user2)
    if err != nil {
        tx.Rollback()
        return "", err
    }

    // Recupera l'ID della nuova conversazione
    newConvID, err := result.LastInsertId()
    if err != nil {
        tx.Rollback()
        return "", err
    }

    // Conferma la transazione
    if err := tx.Commit(); err != nil {
        return "", err
    }

    // Restituisci l'ID della nuova conversazione
    return fmt.Sprintf("%d", newConvID),  nil
}

// IsUserInConversation verifica se un utente è membro di una conversazione
func (db *appdbimpl) IsUserInConversation(userID, convID string) (bool, error) {
    // Controlla il tipo di conversazione
    isPrivate, err := db.IsConversationPrivate(convID)
    if err != nil {
        return false, err
    }

    // Se la conversazione è privata, controlla se l'utente è tra il creatore o l'altro utente
    if isPrivate {
        // Esegui la query SQL per verificare se l'utente è membro della conversazione privata
        var creatorID, otherUserID string
        err := db.c.QueryRow(`
            SELECT creator_id, otherUser FROM conversations WHERE id = ? AND type = 'private'
        `, convID).Scan(&creatorID, &otherUserID)

        // Restituisci un errore se la query fallisce
        if err != nil {
            return false, err
        }

        // Controlla se l'utente è il creatore o l'altro utente della conversazione
        if userID == creatorID || userID == otherUserID {
            return true, nil
        }

        return false, nil
    } else {
        // Esegui la query SQL per verificare se l'utente è membro della conversazione di gruppo
        var exists bool
        err := db.c.QueryRow(`
            SELECT EXISTS (
                SELECT 1 FROM group_members WHERE conversation_id = ? AND user_id = ?
            )
        `, convID, userID).Scan(&exists)

        // Restituisci un errore se la query fallisce
        if err != nil {
            return false, err
        }
        
        return exists, nil
    }

    
}

// ConversationExists verifica se una conversazione esiste nel database
func (db *appdbimpl) ConversationExists(convID string) (bool, error) {
    var exists bool
    err := db.c.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM conversations WHERE id = ?
        )
    `, convID).Scan(&exists)
    return exists, err
}

// isConversationPrivate verifica se una conversazione è di tipo privato
func (db *appdbimpl) IsConversationPrivate(convID string) (bool, error) {
    var isPrivate bool
    err := db.c.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM conversations WHERE id = ? AND type = 'private'
        )
    `, convID).Scan(&isPrivate)

    if err != nil {
        return false, err
    }

    return isPrivate, nil
}

// isUserCreatorOfConversation verifica se un utente è il creatore di una conversazione
func (db *appdbimpl) IsUserCreatorOfGroup(userID, convID string) (bool, error) {
    isPrivate, err := db.IsConversationPrivate(convID)
    if err != nil {
        return false, err
    }
    if !isPrivate {
        var isCreator bool
        err := db.c.QueryRow(`
            SELECT EXISTS (
                SELECT 1 FROM conversations WHERE id = ? AND creator_id = ?
            )
        `, convID, userID).Scan(&isCreator)

        if err != nil {
            return false, err
        }

        return isCreator, nil
    } else {
        return false, fmt.Errorf("400: conversation is private")
    }
}
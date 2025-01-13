package database

import (
    "fmt"
)

// GetUserConversations retrieves all conversations associated with a given user ID.
// It joins the 'conversations' and 'conversation_members' tables to find conversations
// where the specified user is a member.
func (db *appdbimpl) GetUserConversations(userID string) ([]Conversation, error) {

    // Esegui la query SQL per recuperare le conversazioni dell'utente
    rows, err := db.c.Query(`
        SELECT c.id, cm.nickname, c.type, c.creator_id, cm.photo, m.content AS lastMessage
        FROM conversations c
        JOIN conversation_members cm ON c.id = cm.conversation_id
        LEFT JOIN messages m ON c.lastMessageId = m.id
        WHERE cm.user_id = ?`, userID)

    // Restituisci un errore se la query fallisce
    if err != nil {
        return nil, err 
    }

    // Chiudi le righe quando la funzione termina
    defer rows.Close() 

    // Crea una slice per contenere le conversazioni
    var conversations []Conversation

    //  Scansiona le righe restituite dalla query
    for rows.Next() {

        // Crea una variabile di tipo Conversation
        var conv Conversation

        // Scansiona i valori delle colonne nella struttura Conversation
        if err := rows.Scan(&conv.ConvID, &conv.Name, &conv.Type, &conv.CreatorID, &conv.Photo, &conv.LastMessage); err != nil {
            return nil, err
        }

        // Aggiungi la conversazione alla slice
        conversations = append(conversations, conv)
    }

    // Se non ci sono conversazioni, restituisci una slice vuota
    if conversations == nil {
        conversations = []Conversation{}
    }

    // Restituisci la slice di conversazioni
    return conversations, nil 
}

// GetConversationByID retrieves a specific conversation from the database by its ID.
func (db *appdbimpl) GetConversationByID(convID string) (Conversation, error) {
    var conv Conversation // Create a Conversation struct to hold the result

    // Execute the SQL query to get the conversation details by ID
    err := db.c.QueryRow(`
        SELECT id, name, type, creator_id
        FROM conversations
        WHERE id = ?`, convID).Scan(&conv.ConvID, &conv.Name, &conv.Type, &conv.CreatorID)

    if err != nil {
        return conv, err // Return an error if the query fails or no row is found
    }

    return conv, nil // Return the conversation details
}

// DeleteConversation deletes a conversation from the database by its ID.
func (db *appdbimpl) DeleteConversation(convID string) error {
    // Execute the SQL command to delete the conversation by ID
    _, err := db.c.Exec(`DELETE FROM conversations WHERE id = ?`, convID)
    return err // Return any error that occurs during execution
}

func (db *appdbimpl) CreatePrivateConversation(user1 string, user2 string) (string, error) {

    // Controllo se l'utente sta cercando di creare una conversazione con se stesso
    if user1 == user2 {
        return "", fmt.Errorf("400: cannot create a conversation with yourself")
    }

    // Controlla se la conversazione esiste già e recupera nome e foto
    var convID string
    err := db.c.QueryRow(`
        SELECT c.id FROM conversations c
        JOIN conversation_members cm1 ON c.id = cm1.conversation_id
        JOIN conversation_members cm2 ON c.id = cm2.conversation_id
        WHERE cm1.user_id = ? AND cm2.user_id = ? AND c.type = 'private'
    `, user1, user2).Scan(&convID)

    // Se la conversazione esiste già, restituisci l'ID
    if err == nil {
        return convID, nil 
    }

    // Recupera il nome e la foto dell'altro utente
    var user2Name, user2Photo string
    err = db.c.QueryRow(`SELECT name, photo FROM users WHERE id = ?`, user2).Scan(&user2Name, &user2Photo)
    if err != nil {
        return "", fmt.Errorf("404: user not found")
    }

    // Inizia una transazione per garantire la coerenza
    tx, err := db.c.Begin()
    if err != nil {
        return "", err
    }

    // Inserisce la nuova conversazione nella tabella conversations con il nome dell'altro utente e la sua foto profilo
    result, err := tx.Exec(`
        INSERT INTO conversations (name, type, creator_id, lastMessageId, photo)
        VALUES (?, 'private', ?, NULL, ?)`, "private chat", user1, user2Photo)
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

    // Inserisce il membro creatore della conversazione nella tabella conversation_members con nickname e foto dell'altro utente
    _, err = tx.Exec(`
        INSERT INTO conversation_members (conversation_id, user_id, nickname, photo) VALUES (?, ?, ?, ?)`, newConvID, user1, user2Name, user2Photo)
    if err != nil {
        tx.Rollback()
        return "", err
    }

    // Recupera il nome e la foto dell'utente che ha creato la conversazione
    user1name, _ := db.GetUserByID(user1)
    user1photo, _ := db.GetPhotoByID(user1)

    // Inserisce il secondo membro della conversazione nella tabella conversation_members con il nickname e la foto dell'utente che ha creato la conversazione
    _, err = tx.Exec(`
        INSERT INTO conversation_members (conversation_id, user_id, nickname, photo) VALUES (?, ?, ?, ?)`, newConvID, user2, user1name, user1photo)
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
    var exists bool
    err := db.c.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM conversation_members WHERE user_id = ? AND conversation_id = ?
        )
    `, userID, convID).Scan(&exists)

    if err != nil {
        return false, err
    }

    return exists, nil
}

func (db *appdbimpl) GetOtherUserDetailsInConversation(convID, userID string) (string, string, error) {
    var otherUserName, otherUserPhoto string
    err := db.c.QueryRow(`
        SELECT u.name, u.photo
        FROM users u
        JOIN conversation_members cm ON u.id = cm.user_id
        WHERE cm.conversation_id = ? AND cm.user_id != ?
    `, convID, userID).Scan(&otherUserName, &otherUserPhoto)

    if err != nil {
        return "", "", err
    }

    return otherUserName, otherUserPhoto, nil
}

func (db *appdbimpl) ConversationExists(convID string) (bool, error) {
    var exists bool
    err := db.c.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM conversations WHERE id = ?
        )
    `, convID).Scan(&exists)

    if err != nil {
        return false, err
    }

    return exists, nil
}
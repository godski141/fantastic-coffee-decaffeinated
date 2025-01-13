package database

import (
    "fmt"
    "log"
)

// GetUserConversations retrieves all conversations associated with a given user ID.
// It joins the 'conversations' and 'conversation_members' tables to find conversations
// where the specified user is a member.
func (db *appdbimpl) GetUserConversations(userID string) ([]Conversation, error) {
    // Execute the SQL query to get the conversations associated with the user from the tables 'conversations' and 'conversation_members'
    rows, err := db.c.Query(`
        SELECT c.id, c.name, c.type, c.creator_id, c.photo, m.content AS lastMessage
        FROM conversations c
        JOIN conversation_members cm ON c.id = cm.conversation_id
        LEFT JOIN messages m ON c.lastMessageId = m.id
        WHERE cm.user_id = ?`, userID)
    if err != nil {
        return nil, err // Return an error if the query fails
    }
    defer rows.Close() // Ensure rows are closed when the function exits

    // Create a slice to hold the conversations
    var conversations []Conversation

    // Iterate over the rows and scan the values into a Conversation struct
    for rows.Next() {
        var conv Conversation
        // Scan the values from the row into the Conversation struct
        if err := rows.Scan(&conv.ID, &conv.Name, &conv.Type, &conv.CreatorID, &conv.Photo, &conv.LastMessage); err != nil {
            return nil, err // Return an error if scanning fails
        }
        conversations = append(conversations, conv) // Add the conversation to the slice
    }

    return conversations, nil // Return the list of conversations
}

// GetConversationByID retrieves a specific conversation from the database by its ID.
func (db *appdbimpl) GetConversationByID(convID string) (Conversation, error) {
    var conv Conversation // Create a Conversation struct to hold the result

    // Execute the SQL query to get the conversation details by ID
    err := db.c.QueryRow(`
        SELECT id, name, type, creator_id
        FROM conversations
        WHERE id = ?`, convID).Scan(&conv.ID, &conv.Name, &conv.Type, &conv.CreatorID)

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

func (db *appdbimpl) CreatePrivateConversation(user1 string, user2 string) (string, string, string, error) {
    // Controllo se l'utente sta cercando di creare una conversazione con se stesso
    if user1 == user2 {
        return "", "", "", fmt.Errorf("400: cannot create a conversation with yourself")
    }

    // Controlla se la conversazione esiste già e recupera nome e foto
    var convID, convName, convPhoto string
    err := db.c.QueryRow(`
        SELECT c.id, c.name, c.photo FROM conversations c
        JOIN conversation_members cm1 ON c.id = cm1.conversation_id
        JOIN conversation_members cm2 ON c.id = cm2.conversation_id
        WHERE cm1.user_id = ? AND cm2.user_id = ? AND c.type = 'private'
    `, user1, user2).Scan(&convID, &convName, &convPhoto)

    if err == nil {
        return convID, convName, convPhoto, nil // La conversazione esiste già, restituiamo i dettagli
    }

    // Recupera il nome e la foto dell'utente 2
    var user2Name, user2Photo string
    err = db.c.QueryRow(`SELECT name, photo FROM users WHERE id = ?`, user2).Scan(&user2Name, &user2Photo)
    log.Println("DEBUG: ERROR:", err)
    if err != nil {
        return "", "", "", fmt.Errorf("404: user not found")
    }

    // Inizia una transazione per garantire la coerenza
    tx, err := db.c.Begin()
    if err != nil {
        return "", "", "", err
    }

    // Inserisce la nuova conversazione nella tabella conversations con il nome dell'altro utente e la sua foto profilo
    result, err := tx.Exec(`
        INSERT INTO conversations (name, type, creator_id, lastMessageId, photo)
        VALUES (?, 'private', ?, NULL, ?)`, user2Name, user1, user2Photo)
    if err != nil {
        tx.Rollback()
        return "", "", "", err
    }

    // Recupera l'ID della nuova conversazione
    newConvID, err := result.LastInsertId()
    if err != nil {
        tx.Rollback()
        return "", "", "", err
    }

    // Inserisce i membri della conversazione nella tabella conversation_members
    _, err = tx.Exec(`
        INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?)`, newConvID, user1)
    if err != nil {
        tx.Rollback()
        return "", "", "", err
    }

    _, err = tx.Exec(`
        INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?)`, newConvID, user2)
    if err != nil {
        tx.Rollback()
        return "", "", "", err
    }

    // Conferma la transazione
    if err := tx.Commit(); err != nil {
        return "", "", "", err
    }

    return fmt.Sprintf("%d", newConvID), user2Name, user2Photo, nil
}


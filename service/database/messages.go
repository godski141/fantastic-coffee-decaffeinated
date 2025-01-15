package database

import (
	"database/sql"
	"errors"
)

//	InsertMessage inserisce un messaggio nel database
func (db *appdbimpl) InsertMessage(convID string, userID string, text string) (string, error) {

	// Inserisce il messaggio nel database
    var messageID string
    err := db.c.QueryRow(
        "INSERT INTO messages (conversation_id, sender_id, content, status) VALUES (?, ?, ?, 'sent') RETURNING id",
        convID, userID, text,
    ).Scan(&messageID)

	// Restituisce un errore se la query non è andata a buon fine
    if err != nil {
        return "", err
    }

	// Aggiorna l'ultimo messaggio della conversazione e restituisce l'ID del messaggio
    return messageID, nil
}

//	GetMessageFromID recupera un messaggio dal database dato il suo ID
func (db *appdbimpl) GetMessageFromID(messageID string) (Message, error) {
	var message Message
	
	// Esegue la query per recuperare il messaggio
	err := db.c.QueryRow(
		"SELECT id, conversation_id, sender_id, content, timestamp, status FROM messages WHERE id = ?",
		messageID,
	).Scan(&message.MessageID, &message.ConversationID, &message.SenderID, &message.Content, &message.Timestamp, &message.Status)
	
	if err != nil {
		return Message{}, err
	}
	return message, nil
}

// UpdateLastMessage aggiorna l'ultimo messaggio di una conversazione
func (db *appdbimpl) UpdateLastMessage(convID string, messageID string) error {
	_, err := db.c.Exec(
		"UPDATE conversations SET lastMessageId = ? WHERE id = ?",
		messageID, convID,
	)
	return err
}

// MessageExists controlla se un messaggio esiste nel database
func (db *appdbimpl) MessageExists(messageID string) (bool, error) {
	var exists bool
	err := db.c.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM messages WHERE id = ?)",
		messageID,
	).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetSenderIDFromMessageID recupera l'ID del mittente di un messaggio dato il suo ID
func (db *appdbimpl) GetSenderIDFromMessageID(messageID string) (string, error) {
	var senderID string
	err := db.c.QueryRow(
		"SELECT sender_id FROM messages WHERE id = ?",
		messageID,
	).Scan(&senderID)

	if err != nil {
		return "", err
	}

	return senderID, nil
}

// DeleteMessage elimina un messaggio dal database
func (db *appdbimpl) DeleteMessage(messageID string) error {
	_, err := db.c.Exec(
		"DELETE FROM messages WHERE id = ?",
		messageID,
	)
	return err
}

// GetLastMessageID recupera l'ID dell'ultimo messaggio di una conversazione
func (db *appdbimpl) GetLastMessageID(convID string) (string, error) {
    var lastMessageID sql.NullString
    err := db.c.QueryRow(
        "SELECT id FROM messages WHERE conversation_id = ? ORDER BY timestamp DESC LIMIT 1",
        convID,
    ).Scan(&lastMessageID)

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return "", nil // Nessun messaggio rimasto
        }
        return "", err
    }
    return lastMessageID.String, nil
}

// InsertReaction inserisce una reazione a un messaggio nel database
func (db *appdbimpl) InsertReaction(userID string, messageID string, reaction string) error {
    // Controlla se l'utente ha già reagito a questo messaggio
    var exists bool
    err := db.c.QueryRow(
        "SELECT EXISTS(SELECT 1 FROM reactions WHERE message_id = ? AND user_id = ?)",
        messageID, userID,
    ).Scan(&exists)
    if err != nil {
        return err
    }

    // Se esiste già una reazione, la rimuove prima di aggiungere la nuova
    if exists {
        _, err = db.c.Exec(
            "DELETE FROM reactions WHERE message_id = ? AND user_id = ?",
            messageID, userID,
        )
        if err != nil {
            return err
        }
    }

    // Inserisce la nuova reazione
    _, err = db.c.Exec(
        "INSERT INTO reactions (message_id, user_id, reaction) VALUES (?, ?, ?)",
        messageID, userID, reaction,
    )
    if err != nil {
        return err
    }

    // Aggiorna il contatore delle reazioni nel messaggio
    _, err = db.c.Exec(
        "UPDATE messages SET reaction_count = reaction_count + 1 WHERE id = ?",
        messageID,
    )

    return err
}

// DeleteReaction elimina una reazione a un messaggio dal database
func (db *appdbimpl) DeleteReaction(messageID, userID string) error {
    // Elimina la reazione dell'utente per il messaggio specifico
    _, err := db.c.Exec(
        "DELETE FROM reactions WHERE message_id = ? AND user_id = ?",
        messageID, userID,
    )
    if err != nil {
        return err
    }

    // Diminuisce il contatore delle reazioni nel messaggio
    _, err = db.c.Exec(
        "UPDATE messages SET reaction_count = reaction_count - 1 WHERE id = ? AND reaction_count > 0",
        messageID,
    )

    return err
}

// UserHasReaction controlla se un utente ha reagito a un messaggio
func (db *appdbimpl) UserHasReaction(messageID, userID string) (bool, error) {
    var exists bool
    err := db.c.QueryRow(
        "SELECT EXISTS(SELECT 1 FROM reactions WHERE message_id = ? AND user_id = ?)",
        messageID, userID,
    ).Scan(&exists)
    return exists, err
}

// GetMessagesFromConversation recupera tutti i messaggi di una conversazione dal database
func (db *appdbimpl) GetMessagesFromConversation(conversationID string) ([]Message, error) {
    rows, err := db.c.Query(`
        SELECT 
            m.id, m.sender_id, m.content, m.timestamp, m.status,
            COALESCE(r.user_id, '') AS reactionUser, 
            COALESCE(r.reaction, '') AS reaction
        FROM messages m
        LEFT JOIN reactions r ON m.id = r.message_id
        WHERE m.conversation_id = ?
        ORDER BY m.timestamp ASC`, conversationID)

    if err != nil {
        return nil, err
    }
    defer rows.Close()

    messages := make(map[string]Message)

    for rows.Next() {
        var msgID, senderID, content, timestamp, status, reactionUser, reaction string

        if err := rows.Scan(&msgID, &senderID, &content, &timestamp, &status, &reactionUser, &reaction); err != nil {
            return nil, err
        }

        // Se il messaggio non è già nella mappa, lo aggiungiamo
        if _, exists := messages[msgID]; !exists {
            messages[msgID] = Message{
                MessageID: msgID,
                SenderID:  senderID,
                Content:   content,
                Timestamp: timestamp,
                Status:    status,
                Reactions: []Reaction{},
            }
        }

        // Se c'è una reazione, la aggiungiamo
        if reactionUser != "" {
            msg := messages[msgID]
            msg.Reactions = append(msg.Reactions, Reaction{
                UserID:   reactionUser,
                Reaction: reaction,
            })
            messages[msgID] = msg
        }
    }

    // Convertiamo la mappa in una slice di messaggi
    messageList := make([]Message, 0, len(messages))
    for _, msg := range messages {
        messageList = append(messageList, msg)
    }

    return messageList, nil
}


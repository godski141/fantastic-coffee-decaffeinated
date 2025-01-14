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
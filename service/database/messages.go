package database


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
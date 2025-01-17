package database

import "fmt"

func (db *appdbimpl) CreateGroup(name, creatorID string) (string, error) {
	var groupID string
	err := db.c.QueryRow(
		"INSERT INTO conversations (name, creator_id, type, photo, lastMessageId, otherUser) VALUES (?, ?, 'group', '', NULL, NULL) RETURNING id;",
		name, creatorID,
	).Scan(&groupID)
	return groupID, err
}

func (db *appdbimpl) AddUserToGroup(groupID, userID string) error {
	_, err := db.c.Exec(
		"INSERT INTO group_members (conversation_id, user_id) VALUES (?, ?);",
		groupID, userID,
	)
	return err
}

// GetNameFromGroupID restituisce il nome del gruppo con l'id specificato
func (db *appdbimpl) GetNameFromGroupID(groupID string) (string, error) {
	// Controlla che il gruppo sia effettivamente un gruppo
	isPrivate, err := db.IsConversationPrivate(groupID)
	if err != nil {
		return "", err
	}
	if isPrivate {
		return "", fmt.Errorf("404: Group not found")
	}
	var name string
	err2 := db.c.QueryRow("SELECT name FROM conversations WHERE id = ? ", groupID).Scan(&name)
	return name, err2
}
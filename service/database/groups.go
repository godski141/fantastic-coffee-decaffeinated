package database

import (
	"database/sql"
	"fmt"
)

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

// ChangeGroupName cambia il nome del gruppo con l'id specificato
func (db *appdbimpl) ChangeGroupName(groupID, name string) error {
	// Esegue la query per cambiare il nome del gruppo
	_, err := db.c.Exec("UPDATE conversations SET name = ? WHERE id = ?", name, groupID)
	return err
}

// leaveGroup rimuove l'utente con l'id specificato dal gruppo con l'id specificato
func (db *appdbimpl) LeaveGroup(groupID, userID string) error {
	// Esegue la query per rimuovere l'utente dal gruppo
	_, err := db.c.Exec("DELETE FROM group_members WHERE conversation_id = ? AND user_id = ?", groupID, userID)
	return err
}

// getGroupPhotoByID restituisce la foto del gruppo con l'id specificato
func (db *appdbimpl) GetGroupPhotoByID(groupID string) (string, error) {
	var photo sql.NullString
	err2 := db.c.QueryRow("SELECT photo FROM conversations WHERE id = ? ", groupID).Scan(&photo)
	if err2 != nil {
		return "", err2
	}
	if photo.Valid {
		return photo.String, nil
	} else {
		return "", nil
	}
}

// UpdateGroupPhoto cambia la foto del gruppo con l'id specificato
func (db *appdbimpl) UpdateGroupPhoto(groupID, photoPath string) error {
	// Esegue la query per cambiare la foto del gruppo
	_, err := db.c.Exec("UPDATE conversations SET photo = ? WHERE id = ?", photoPath, groupID)
	return err
}
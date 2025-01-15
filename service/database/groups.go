package database

func (db *appdbimpl) CreateGroup(name, creatorID string) (string, error) {
    var groupID string
    err := db.c.QueryRow(
        "INSERT INTO conversations (name, creator_id, type) VALUES (?, ?, 'group') RETURNING id;",
        name, creatorID,
    ).Scan(&groupID)
    return groupID, err
}


func (db *appdbimpl) AddUserToGroup(groupID, userID string) error {
    name, _ := db.GetNameFromGroupID(groupID)
    _, err := db.c.Exec(
        "INSERT INTO conversation_members (conversation_id, user_id, nickname, photo) VALUES (?, ?, ?, '');",
        groupID, userID, name,
    )
    return err
}


// GetNameFromGroupID restituisce il nome del gruppo con l'id specificato
func (db *appdbimpl) GetNameFromGroupID(groupID string) (string, error) {
    var name string
    err := db.c.QueryRow("SELECT name FROM conversations WHERE id = ?", groupID).Scan(&name)
    return name, err
}
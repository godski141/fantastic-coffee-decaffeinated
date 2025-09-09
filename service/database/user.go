package database

import (
	"database/sql"
	"fmt"
	"log"
)

// CreateUser crea un nuovo utente con il nome specificato
func (db *appdbimpl) CreateUser(name string) (string, error) {
    var id string
    err := db.c.QueryRow("INSERT INTO users (name, photo) VALUES (?, '') RETURNING id", name).Scan(&id)
    return id, err
}

// GetUserByID restituisce il nome dell'utente con l'id specificato
func (db *appdbimpl) GetUserByID(id string) (string, error) {
    var name string
    err := db.c.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&name)
    return name, err
}

// GetPhotoByID restituisce la foto dell'utente con l'id specificato
func (db *appdbimpl) GetUserPhotoByID(id string) (string, error) {
    var photo sql.NullString
	err := db.c.QueryRow("SELECT photo FROM users WHERE id = ? ", id).Scan(&photo)
	if err != nil {
        return "", err
    }
    if photo.String == "" {
        return "", nil
    }
    
    return photo.String, nil
}

// GetUserByName restituisce l'id dell'utente con il nome specificato
func (db *appdbimpl) GetUserByName(name string) (string, error) {
    var id string
    log.Println("DEBUG: Searching for user: ", name)
    err := db.c.QueryRow("SELECT id FROM users WHERE name = ?", name).Scan(&id)
    if err != nil{
        log.Println("ERROR: User not found in database:", name)
        return "", fmt.Errorf("404: User not found")
    }
    log.Println("DEBUG: Found user: ", name, "with id: ", id)
    return id, nil
}

// ModifyUserName modifica il nome dell'utente con l'id specificato
func (db *appdbimpl) ModifyUserName(id string, name string) error {
    _, err := db.c.Exec("UPDATE users SET name = ? WHERE id = ?", name, id)
    return err
}

// SearchUsersByUsername cerca utenti per username parziale
func (db *appdbimpl) SearchUsersByUsername(query string) ([]string, error) {
    // Query per cercare utenti che contengono la stringa di ricerca
    // Utilizziamo LIKE con % per ricerca parziale, limitiamo a 100 risultati
    rows, err := db.c.Query("SELECT name FROM users WHERE name LIKE ? ORDER BY name LIMIT 100", "%"+query+"%")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var usernames []string
    for rows.Next() {
        var username string
        if err := rows.Scan(&username); err != nil {
            return nil, err
        }
        usernames = append(usernames, username)
    }
    
    // Controlla se ci sono stati errori durante l'iterazione
    if err := rows.Err(); err != nil {
        return nil, err
    }
    
    return usernames, nil
}

// updateUserPhoto aggiorna la foto dell'utente con l'id specificato
func (db *appdbimpl) UpdateUserPhoto(id string, photoPath string) error {
    _, err := db.c.Exec("UPDATE users SET photo = ? WHERE id = ?", photoPath,
        id)
    return err
}
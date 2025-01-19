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

// updateUserPhoto aggiorna la foto dell'utente con l'id specificato
func (db *appdbimpl) UpdateUserPhoto(id string, photoPath string) error {
    _, err := db.c.Exec("UPDATE users SET photo = ? WHERE id = ?", photoPath,
        id)
    return err
}
package database

import (
    "fmt"
    "log"
)

func (db *appdbimpl) CreateUser(name string) (string, error) {
    var id string
    err := db.c.QueryRow("INSERT INTO users (name, photo) VALUES (?, '') RETURNING id", name).Scan(&id)
    return id, err
}

func (db *appdbimpl) GetUserByID(id string) (string, error) {
    var name string
    err := db.c.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&name)
    return name, err
}

func (db *appdbimpl) GetPhotoByID(id string) (string, error) {
    var photo string
    err := db.c.QueryRow("SELECT photo FROM users WHERE id = ?", id).Scan(&photo)
    return photo, err
}

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
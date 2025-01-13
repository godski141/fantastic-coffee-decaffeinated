/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	GetName() (string, error)
	SetName(name string) error

	CreateUser(name string) (string, error)
	GetUserByID(id string) (string, error)
    GetPhotoByID(id string) (string, error)
	GetUserByName(name string) (string, error)
    

    GetUserConversations(userID string) ([]Conversation, error)
    GetConversationByID(convID string) (Conversation, error)
    DeleteConversation(convID string) error
    CreatePrivateConversation(user1 string, user2 string) (string, error)
	
	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
    if db == nil {
        return nil, errors.New("database is required when building a AppDatabase")
    }

    // Check if tables exist. If not, the database is empty, and we need to create the structure
    tables := []string{"users", "conversations", "messages", "conversation_members"}
    for _, table := range tables {
        var tableName string
        err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?;`, table).Scan(&tableName)
        if errors.Is(err, sql.ErrNoRows) {
            var sqlStmt string
            switch table {
            case "users":
                sqlStmt = `CREATE TABLE users (
                    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
                    name TEXT NOT NULL UNIQUE,
					photo TEXT 
                );`
            case "conversations":
                sqlStmt = `CREATE TABLE conversations (
                    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
                    name TEXT NOT NULL,
                    type TEXT CHECK(type IN ('private', 'group')) NOT NULL,
                    creator_id INTEGER NOT NULL,
					photo TEXT,
                    lastMessageId INTEGER,
                    FOREIGN KEY (creator_id) REFERENCES users(id),
                    FOREIGN KEY (lastMessageId) REFERENCES messages(id) ON DELETE SET NULL
                );`
            case "messages":
                sqlStmt = `CREATE TABLE messages (
                    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
                    conversation_id INTEGER NOT NULL,
                    sender_id INTEGER NOT NULL,
                    content TEXT NOT NULL,
                    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                    status TEXT CHECK(status IN ('sent', 'received', 'read')) NOT NULL,
                    FOREIGN KEY (conversation_id) REFERENCES conversations(id),
                    FOREIGN KEY (sender_id) REFERENCES users(id)
                );`
            case "conversation_members":
                sqlStmt = `CREATE TABLE conversation_members (
                    conversation_id INTEGER NOT NULL,
                    user_id INTEGER NOT NULL,
                    nickname TEXT,
                    photo TEXT,
                    FOREIGN KEY (conversation_id) REFERENCES conversations(id),
                    FOREIGN KEY (user_id) REFERENCES users(id),
                    PRIMARY KEY (conversation_id, user_id)
                );`
            }
            _, err = db.Exec(sqlStmt)
            if err != nil {
                return nil, fmt.Errorf("error creating table %s: %w", table, err)
            }
        }
    }

    return &appdbimpl{
        c: db,
    }, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

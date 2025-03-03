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
    GetUserPhotoByID(id string) (string, error)
	GetUserByName(name string) (string, error)
    ModifyUserName(id string, name string) error
    UpdateUserPhoto(id string, photoPath string) error

    GetUserConversations(userID string) ([]Conversation, error)
    GetConversationByID(convID, userID string) (Conversation, error)
    DeleteConversation(convID string) error
    CreatePrivateConversation(user1 string, user2 string) (string, error)
    IsUserInConversation(userID, convID string) (bool, error)
    ConversationExists(convID string) (bool, error)
    GetMessagesFromConversation(conversationID string) ([]Message, error)
    IsConversationPrivate(convID string) (bool, error)
    IsUserCreatorOfGroup(userID, convID string) (bool, error)
	
    InsertMessage(convID string, userID string, text string) (string, error)
    GetMessageFromID(messageID string) (Message, error)
    UpdateLastMessage(convID string, messageID string) error
    MessageExists(messageID string) (bool, error)
    DeleteMessage(messageID string) error
    GetLastMessageID(convID string) (string, error)
    InsertReaction(messageID string, userID string, reaction string) error
    DeleteReaction(messageID, userID string) error
    UserHasReaction(messageID, userID string) (bool, error)
    GetContentFromMessageID(messageID string) (string, error)

    CreateGroup(name, creatorID string) (string, error)
    AddUserToGroup(groupID, userID string) error
    ChangeGroupName(groupID, name string) error
    LeaveGroup(groupID, userID string) error
    GetGroupPhotoByID(groupID string) (string, error)
    UpdateGroupPhoto(groupID, photoPath string) error

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
    
    // Enable foreign keys
	_, err := db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

    // Check if tables exist. If not, the database is empty, and we need to create the structure
    tables := []string{"users", "conversations", "messages", "group_members", "reactions"}
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
                    name TEXT,
                    type TEXT CHECK(type IN ('private', 'group')) NOT NULL,
                    creator_id INTEGER NOT NULL,
					photo TEXT,
                    lastMessageId INTEGER,
                    otherUser INTEGER,
                    FOREIGN KEY (creator_id) REFERENCES users(id),
                    FOREIGN KEY (lastMessageId) REFERENCES messages(id) ON DELETE SET NULL,
                    FOREIGN KEY (otherUser) REFERENCES users(id) ON DELETE SET NULL
                );`
            case "messages":
                sqlStmt = `CREATE TABLE messages (
                    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
                    conversation_id INTEGER NOT NULL,
                    sender_id INTEGER NOT NULL,
                    content TEXT NOT NULL,
                    reaction_count INTEGER DEFAULT 0,
                    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                    status TEXT CHECK(status IN ('sent', 'received', 'read')) NOT NULL,
                    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
                    FOREIGN KEY (sender_id) REFERENCES users(id)
                );`
            case "group_members":
                sqlStmt = `CREATE TABLE group_members (
                    conversation_id INTEGER NOT NULL,
                    user_id INTEGER NOT NULL,
                    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
                    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
                    PRIMARY KEY (conversation_id, user_id)
                );`
            case "reactions":
                sqlStmt = `CREATE TABLE reactions (
                    message_id INTEGER NOT NULL,
                    user_id INTEGER NOT NULL,
                    reaction TEXT NOT NULL,
                    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                    PRIMARY KEY (message_id, user_id), 
                    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
                    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
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

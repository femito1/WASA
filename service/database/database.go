/*
Package database is the middleware between the app database and the code.
All data (de)serialization (save/load) from a persistent database are handled here.
Database-specific logic should never escape this package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrUserDoesNotExist = errors.New("User does not exist")

// User represents a user in the system.
type User struct {
	Id             uint64 `json:"userId"`
	Username       string `json:"name"`
	ProfilePicture string `json:"profilePicture,omitempty"`
}

// AppDatabase is the high level interface for the DB.
type AppDatabase interface {
	// User operations
	CreateUser(User) (User, error)
	SetUsername(User, string) (User, error)
	SetPhoto(User, string) (User, error)
	GetUserId(string) (User, error)
	CheckUserById(User) (User, error)
	CheckUserByUsername(User) (User, error)
	CheckUser(User) (User, error)
	ListUsers(nameFilter string) ([]User, error)
	Ping() error

	// Conversation operations
	CreateConversation(creator User, convName string, members []uint64) (Conversation, error)
	GetConversations(userId uint64) ([]Conversation, error)
	GetConversation(userId, convId uint64, conversationName *string) (Conversation, error)
	SetConversationName(userId, convId uint64, newName string) (Conversation, error)
	SetConversationPhoto(userId, convId uint64, newPhoto string) (Conversation, error)
	AddUserToConversation(userId, convId, userIdToAdd uint64) (Conversation, error)
	RemoveUserFromConversation(userId, convId uint64) error

	// Message operations
	CreateMessage(sender User, convId uint64, content string, format string, replyTo *uint64) (Message, error)
	DeleteMessage(user User, convId, msgId uint64) error
	ForwardMessage(user User, convId, msgId, targetConvId uint64) (Message, error)
	GetMessageByID(msgId uint64) (Message, error)

	// Comment operations
	CommentMessage(user User, convId, msgId uint64, commentText string) (uint64, error)
	DeleteComment(user User, convId, msgId, commentId uint64) error

	// Contacts
	AddContact(userID, contactID uint64) error
	ListContacts(userID uint64) ([]User, error)
	RemoveContact(userID, contactID uint64) error

	// Message reactions
	ReactToMessage(user User, convId, msgId uint64, emoji string) error
}

// appdbimpl is the concrete implementation of AppDatabase.
type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// It also creates (or migrates) all necessary tables.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building an AppDatabase")
	}

	// Enable foreign keys
	_, err := db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	// Create necessary tables.
	// Users table (includes profilePicture)
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		profilePicture TEXT
	);
	`
	if _, err := db.Exec(usersTable); err != nil {
		return nil, fmt.Errorf("error creating users table: %w", err)
	}

	// Conversations table.
	conversationsTable := `
	CREATE TABLE IF NOT EXISTS conversations (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		picture TEXT
	);
	`
	if _, err := db.Exec(conversationsTable); err != nil {
		return nil, fmt.Errorf("error creating conversations table: %w", err)
	}

	// Conversation members (many-to-many).
	conversationMembersTable := `
	CREATE TABLE IF NOT EXISTS conversation_members (
		conversation_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		PRIMARY KEY (conversation_id, user_id),
		FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`
	if _, err := db.Exec(conversationMembersTable); err != nil {
		return nil, fmt.Errorf("error creating conversation_members table: %w", err)
	}

	// Messages table.
	messagesTable := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		conversation_id INTEGER NOT NULL,
		sender_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		format TEXT NOT NULL,
		state TEXT NOT NULL,
		is_forwarded INTEGER NOT NULL DEFAULT 0,
		reply_to INTEGER,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
		FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`
	if _, err := db.Exec(messagesTable); err != nil {
		return nil, fmt.Errorf("error creating messages table: %w", err)
	}

	// Reactions table.
	reactionsTable := `
	CREATE TABLE IF NOT EXISTS reactions (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		message_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		emoji TEXT NOT NULL,
		UNIQUE(message_id, user_id),
		FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`
	if _, err := db.Exec(reactionsTable); err != nil {
		return nil, fmt.Errorf("error creating reactions table: %w", err)
	}

	// Comments table.
	commentsTable := `
	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		message_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		commentText TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`
	if _, err := db.Exec(commentsTable); err != nil {
		return nil, fmt.Errorf("error creating comments table: %w", err)
	}
	// In database/database.go (inside New(db *sql.DB))
	const contactsTable = `
	CREATE TABLE IF NOT EXISTS contacts (
	user_id INTEGER NOT NULL,
	contact_id INTEGER NOT NULL,
	PRIMARY KEY (user_id, contact_id),
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY (contact_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`
	if _, err := db.Exec(contactsTable); err != nil {
		return nil, fmt.Errorf("error creating contacts table: %w", err)
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

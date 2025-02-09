package database

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrConversationNotFound = errors.New("conversation not found")

// getConversationMembers retrieves the members of a conversation.
func (db *appdbimpl) getConversationMembers(convId uint64) ([]User, error) {
	query := `
	SELECT u.id, u.username, u.profilePicture
	FROM users u
	INNER JOIN conversation_members cm ON u.id = cm.user_id
	WHERE cm.conversation_id = ?
	`
	rows, err := db.c.Query(query, convId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Id, &u.Username, &u.ProfilePicture); err != nil {
			return nil, err
		}
		members = append(members, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

func (db *appdbimpl) getConversationMessages(convId uint64) ([]Message, error) {
	query := `
		SELECT m.id, m.sender_id, u.username, m.content, m.format, m.state, m.timestamp
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = ?
		ORDER BY m.timestamp ASC
	`
	rows, err := db.c.Query(query, convId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		err = rows.Scan(&m.Id, &m.SenderId, &m.SenderName, &m.Content, &m.Format, &m.State, &m.Timestamp)
		if err != nil {
			return nil, err
		}
		// Load reactions (if applicable).
		m.Reactions, err = db.getMessageReactions(m.Id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		messages = append(messages, m)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return messages, nil
}

// CreateConversation creates a new conversation with an optional initial set of members.
func (db *appdbimpl) CreateConversation(creator User, convName string, members []uint64) (Conversation, error) {
	var conv Conversation

	// Create the conversation.
	res, err := db.c.Exec("INSERT INTO conversations(name, picture) VALUES (?, ?)", convName, "")
	if err != nil {
		return conv, err
	}
	convId, err := res.LastInsertId()
	if err != nil {
		return conv, err
	}
	conv.Id = uint64(convId)
	conv.Name = convName
	conv.Picture = ""

	// IMPORTANT: Always add the creator to the conversation members.
	_, err = db.c.Exec("INSERT INTO conversation_members(conversation_id, user_id) VALUES (?, ?)", conv.Id, creator.Id)
	if err != nil {
		return conv, err
	}

	// Add additional members, if any.
	for _, uid := range members {
		_, err = db.c.Exec("INSERT OR IGNORE INTO conversation_members(conversation_id, user_id) VALUES (?, ?)", conv.Id, uid)
		if err != nil {
			return conv, err
		}
	}

	// Load the conversation members.
	conv.Members, err = db.getConversationMembers(conv.Id)
	if err != nil {
		return conv, err
	}

	// Initialize the messages slice.
	conv.Messages = []Message{}

	return conv, nil
}

// GetConversations returns all conversations in which the user is a member.
func (db *appdbimpl) GetConversations(userId uint64) ([]Conversation, error) {
	query := `
	SELECT c.id, c.name, c.picture
	FROM conversations c
	INNER JOIN conversation_members cm ON c.id = cm.conversation_id
	WHERE cm.user_id = ?
	`
	rows, err := db.c.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []Conversation
	for rows.Next() {
		var conv Conversation
		if err := rows.Scan(&conv.Id, &conv.Name, &conv.Picture); err != nil {
			return nil, err
		}
		// Load members and messages.
		conv.Members, err = db.getConversationMembers(conv.Id)
		if err != nil {
			return nil, err
		}
		conv.Messages, err = db.getConversationMessages(conv.Id)
		if err != nil {
			return nil, err
		}
		convs = append(convs, conv)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return convs, nil
}

// GetConversation retrieves a specific conversation if the user is a member.
// If conversationName is provided, it must match the conversation's name.
func (db *appdbimpl) GetConversation(userID, convID uint64, conversationName *string) (Conversation, error) {
	var conv Conversation

	// Retrieve the conversation info.
	query := "SELECT id, name, picture FROM conversations WHERE id = ?"
	row := db.c.QueryRow(query, convID)
	if err := row.Scan(&conv.Id, &conv.Name, &conv.Picture); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return conv, ErrConversationNotFound
		}
		return conv, fmt.Errorf("error scanning conversation: %w", err)
	}

	// Optionally, verify the conversation name if provided.
	if conversationName != nil && conv.Name != *conversationName {
		return conv, fmt.Errorf("conversation name mismatch")
	}

	// Load conversation members.
	members, err := db.getConversationMembers(convID)
	if err != nil {
		return conv, fmt.Errorf("error loading conversation members: %w", err)
	}
	conv.Members = members

	// Verify that the requesting user is a member.
	memberFound := false
	for _, member := range members {
		if member.Id == userID {
			memberFound = true
			break
		}
	}
	if !memberFound {
		return conv, errors.New("user is not a member of this conversation")
	}

	// Load conversation messages.
	messages, err := db.getConversationMessages(convID)
	if err != nil {
		return conv, fmt.Errorf("error loading conversation messages: %w", err)
	}
	conv.Messages = messages

	return conv, nil
}

// SetConversationName updates the name of a conversation.
func (db *appdbimpl) SetConversationName(userId, convId uint64, newName string) (Conversation, error) {
	// Verify membership.
	var count int
	err := db.c.QueryRow("SELECT COUNT(*) FROM conversation_members WHERE conversation_id = ? AND user_id = ?", convId, userId).Scan(&count)
	if err != nil {
		return Conversation{}, err
	}
	if count == 0 {
		return Conversation{}, errors.New("user is not a member of the conversation")
	}
	// Update the name.
	_, err = db.c.Exec("UPDATE conversations SET name = ? WHERE id = ?", newName, convId)
	if err != nil {
		return Conversation{}, err
	}
	return db.GetConversation(userId, convId, nil)
}

// SetConversationPhoto updates the photo of a conversation.
func (db *appdbimpl) SetConversationPhoto(userId, convId uint64, newPhoto string) (Conversation, error) {
	// Verify membership.
	var count int
	err := db.c.QueryRow("SELECT COUNT(*) FROM conversation_members WHERE conversation_id = ? AND user_id = ?", convId, userId).Scan(&count)
	if err != nil {
		return Conversation{}, err
	}
	if count == 0 {
		return Conversation{}, errors.New("user is not a member of the conversation")
	}
	// Update the photo.
	_, err = db.c.Exec("UPDATE conversations SET picture = ? WHERE id = ?", newPhoto, convId)
	if err != nil {
		return Conversation{}, err
	}
	return db.GetConversation(userId, convId, nil)
}

// AddUserToConversation adds a new member to an existing conversation.
func (db *appdbimpl) AddUserToConversation(userId, convId, userIdToAdd uint64) (Conversation, error) {
	// Verify that the calling user is a member.
	var count int
	err := db.c.QueryRow("SELECT COUNT(*) FROM conversation_members WHERE conversation_id = ? AND user_id = ?", convId, userId).Scan(&count)
	if err != nil {
		return Conversation{}, err
	}
	if count == 0 {
		return Conversation{}, errors.New("user is not a member of the conversation")
	}
	// Add the new member.
	_, err = db.c.Exec("INSERT OR IGNORE INTO conversation_members(conversation_id, user_id) VALUES (?, ?)", convId, userIdToAdd)
	if err != nil {
		return Conversation{}, err
	}
	return db.GetConversation(userId, convId, nil)
}

// RemoveUserFromConversation removes a user from a conversation.
func (db *appdbimpl) RemoveUserFromConversation(userId, convId uint64) error {
	res, err := db.c.Exec("DELETE FROM conversation_members WHERE conversation_id = ? AND user_id = ?", convId, userId)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("user was not a member of the conversation")
	}
	return nil
}

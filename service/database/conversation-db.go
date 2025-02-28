package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

var ErrConversationNotFound = errors.New("conversation not found")
var logger = logrus.New()

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
	defer func() {
		err := rows.Close()
		if err != nil {
			logger.WithError(err).Error("error closing conversation members rows")
		}
	}()
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
		SELECT m.id, m.sender_id, u.username, u.profilePicture, m.content, m.format, m.state, m.timestamp, m.reply_to, m.is_forwarded
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = ?
		ORDER BY m.timestamp ASC
	`
	rows, err := db.c.Query(query, convId)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			logger.WithError(err).Error("error closing conversation messages rows")
		}
	}()

	var messages []Message
	for rows.Next() {
		var m Message
		var replyTo sql.NullInt64
		var isForwarded int
		err = rows.Scan(&m.Id, &m.SenderId, &m.SenderName, &m.SenderPicture, &m.Content, &m.Format, &m.State, &m.Timestamp, &replyTo, &isForwarded)
		if err != nil {
			return nil, err
		}
		if replyTo.Valid {
			// Load the replied-to message details.
			rMsg, err := db.GetMessageByID(uint64(replyTo.Int64))
			if err != nil {
				return nil, err
			}
			m.ReplyTo = &rMsg
			id := uint64(replyTo.Int64)
			m.ReplyToID = &id
		}
		m.IsForwarded = isForwarded != 0
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

// CreateConversation creates either a direct conversation (2 users) or a group (>2 users).
// For direct conversations, the other user must be in the creator's contacts.
func (db *appdbimpl) CreateConversation(creator User, convName string, members []uint64) (Conversation, error) {
	var conv Conversation

	// Validate members
	if len(members) == 0 {
		return conv, errors.New("cannot create conversation: no other members specified")
	}

	// For direct conversations (just one other member)
	if len(members) == 1 {

		// Check if a direct conversation already exists between these users
		rows, err := db.c.Query(`
			SELECT DISTINCT c.id 
			FROM conversations c
			JOIN conversation_members cm1 ON c.id = cm1.conversation_id
			JOIN conversation_members cm2 ON c.id = cm2.conversation_id
			WHERE cm1.user_id = ? AND cm2.user_id = ?
			AND (
				SELECT COUNT(*) 
				FROM conversation_members cm3 
				WHERE cm3.conversation_id = c.id
			) = 2`, creator.Id, members[0])
		if err != nil {
			return conv, fmt.Errorf("error checking existing conversations: %w", err)
		}
		defer func() {
			err := rows.Close()
			if err != nil {
				logger.WithError(err).Error("error closing conversation rows")
			}
		}()

		if rows.Next() {
			var existingConvId uint64
			if err := rows.Scan(&existingConvId); err != nil {
				return conv, fmt.Errorf("error scanning conversation ID: %w", err)
			}
			return db.GetConversation(creator.Id, existingConvId, nil)
		}
		if rows.Err() != nil {
			return conv, fmt.Errorf("error iterating through existing conversations: %w", rows.Err())
		}
	}

	// Create the conversation
	res, err := db.c.Exec("INSERT INTO conversations(name, picture) VALUES (?, ?)", convName, "")
	if err != nil {
		return conv, fmt.Errorf("error creating conversation: %w", err)
	}

	convId, err := res.LastInsertId()
	if err != nil {
		return conv, fmt.Errorf("error getting conversation ID: %w", err)
	}

	conv.Id = uint64(convId)
	conv.Name = convName
	conv.Picture = ""

	// Add creator as member
	_, err = db.c.Exec("INSERT INTO conversation_members(conversation_id, user_id) VALUES (?, ?)", conv.Id, creator.Id)
	if err != nil {
		return conv, fmt.Errorf("error adding creator to conversation: %w", err)
	}

	// Add other members
	for _, uid := range members {
		_, err = db.c.Exec("INSERT INTO conversation_members(conversation_id, user_id) VALUES (?, ?)", conv.Id, uid)
		if err != nil {
			return conv, fmt.Errorf("error adding member %d: %w", uid, err)
		}
	}

	// Load the conversation members
	conv.Members, err = db.getConversationMembers(conv.Id)
	if err != nil {
		return conv, fmt.Errorf("error getting conversation members: %w", err)
	}

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
	defer func() {
		err := rows.Close()
		if err != nil {
			logger.WithError(err).Error("error closing conversations rows")
		}
	}()

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

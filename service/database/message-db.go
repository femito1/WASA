package database

import (
	"database/sql"
	"errors"
)

// CreateMessage creates a new message in a conversation.
func (db *appdbimpl) CreateMessage(sender User, convId uint64, content string, format string) (Message, error) {
	var msg Message
	// Insert message with state "Sent"
	res, err := db.c.Exec("INSERT INTO messages(conversation_id, sender_id, content, format, state) VALUES (?, ?, ?, ?, ?)",
		convId, sender.Id, content, format, "Sent")
	if err != nil {
		return msg, err
	}
	msgId, err := res.LastInsertId()
	if err != nil {
		return msg, err
	}
	msg.Id = uint64(msgId)
	msg.ConversationId = convId
	msg.SenderId = sender.Id
	msg.Content = content
	msg.Format = format
	msg.State = "Sent"
	// Retrieve timestamp.
	err = db.c.QueryRow("SELECT timestamp FROM messages WHERE id = ?", msg.Id).Scan(&msg.Timestamp)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return msg, err
	}
	msg.Reactions = []Reaction{}
	return msg, nil
}

// DeleteMessage deletes a message from a conversation.
// Only the sender is allowed to delete their message.
func (db *appdbimpl) DeleteMessage(user User, convId, msgId uint64) error {
	res, err := db.c.Exec("DELETE FROM messages WHERE id = ? AND conversation_id = ? AND sender_id = ?", msgId, convId, user.Id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("message not found or user not authorized to delete")
	}
	return nil
}

// ForwardMessage forwards a message to another conversation.
func (db *appdbimpl) ForwardMessage(user User, convId, msgId, targetConvId uint64) error {
	// Retrieve the original message.
	var content, format string
	err := db.c.QueryRow("SELECT content, format FROM messages WHERE id = ? AND conversation_id = ?", msgId, convId).
		Scan(&content, &format)
	if err != nil {
		return err
	}
	// Insert a new message in the target conversation.
	_, err = db.c.Exec("INSERT INTO messages(conversation_id, sender_id, content, format, state) VALUES (?, ?, ?, ?, ?)",
		targetConvId, user.Id, content, format, "Sent")
	return err
}

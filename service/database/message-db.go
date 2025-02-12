package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) CreateMessage(sender User, convId uint64, content string, format string, replyTo *uint64) (Message, error) {
	var msg Message
	// Insert the message with state "Sent" and optional reply_to field.
	res, err := db.c.Exec(
		"INSERT INTO messages(conversation_id, sender_id, content, format, state, reply_to) VALUES (?, ?, ?, ?, ?, ?)",
		convId, sender.Id, content, format, "Sent", replyTo,
	)
	if err != nil {
		return msg, err
	}
	lastInsertId, err := res.LastInsertId()
	if err != nil {
		return msg, err
	}
	msg.Id = uint64(lastInsertId)
	msg.ConversationId = convId
	msg.SenderId = sender.Id
	msg.SenderName = sender.Username
	msg.Content = content
	msg.Format = format
	msg.State = "Sent"
	// Retrieve the timestamp
	err = db.c.QueryRow("SELECT timestamp FROM messages WHERE id = ?", msg.Id).Scan(&msg.Timestamp)
	if err != nil {
		return msg, err
	}
	// Set reply info (replyTo field remains nil here; it will be loaded when fetching conversation messages)
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

// ReactToMessage adds a reaction to a message.
func (db *appdbimpl) ReactToMessage(user User, convId, msgId uint64, emoji string) error {
	// Verify the message exists.
	var dummy int
	err := db.c.QueryRow("SELECT 1 FROM messages WHERE id = ? AND conversation_id = ?", msgId, convId).Scan(&dummy)
	if err != nil {
		return err
	}
	var existing string
	err = db.c.QueryRow("SELECT emoji FROM reactions WHERE message_id = ? AND user_id = ?", msgId, user.Id).Scan(&existing)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No previous reaction found; insert new reaction.
			_, err := db.c.Exec("INSERT INTO reactions(message_id, user_id, emoji) VALUES (?, ?, ?)", msgId, user.Id, emoji)
			return err
		}
		return err
	}
	// Reaction exists, now check if it's the same.
	if existing == emoji {
		// Same reaction already exists; do nothing.
		return nil
	}
	// Update existing reaction to the new emoji.
	_, err = db.c.Exec("UPDATE reactions SET emoji = ? WHERE message_id = ? AND user_id = ?", emoji, msgId, user.Id)
	return err
}

// RemoveReaction removes a reaction from a message.
func (db *appdbimpl) RemoveReaction(user User, convId, msgId uint64, emoji string) error {
	res, err := db.c.Exec("DELETE FROM reactions WHERE message_id = ? AND user_id = ? AND emoji = ?", msgId, user.Id, emoji)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("reaction not found")
	}
	return nil
}

// Modify getMessageReactions to group by emoji and count them.
func (db *appdbimpl) getMessageReactions(msgId uint64) ([]Reaction, error) {
	query := `SELECT emoji, COUNT(*) FROM reactions WHERE message_id = ? GROUP BY emoji`
	rows, err := db.c.Query(query, msgId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []Reaction
	for rows.Next() {
		var r Reaction
		var count int
		if err := rows.Scan(&r.Emoji, &count); err != nil {
			return nil, err
		}
		r.Count = count
		reactions = append(reactions, r)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return reactions, nil
}

// GetMessageByID retrieves a message by its id.
func (db *appdbimpl) GetMessageByID(msgId uint64) (Message, error) {
	var m Message
	var replyTo sql.NullInt64
	query := `
	SELECT m.id, m.sender_id, u.username, m.content, m.format, m.state, m.timestamp, m.reply_to
	FROM messages m
	JOIN users u ON m.sender_id = u.id
	WHERE m.id = ?
	`
	err := db.c.QueryRow(query, msgId).Scan(&m.Id, &m.SenderId, &m.SenderName, &m.Content, &m.Format, &m.State, &m.Timestamp, &replyTo)
	if err != nil {
		return m, err
	}
	if replyTo.Valid {
		id := uint64(replyTo.Int64)
		m.ReplyToID = &id
	}
	return m, nil
}

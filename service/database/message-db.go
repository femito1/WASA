package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) CreateMessage(sender User, convId uint64, content string, format string, replyTo *uint64) (Message, error) {
	res, err := db.c.Exec(
		"INSERT INTO messages(conversation_id, sender_id, content, format, state, reply_to) VALUES (?, ?, ?, ?, ?, ?)",
		convId, sender.Id, content, format, "Sent", replyTo,
	)
	if err != nil {
		var empty Message
		return empty, err
	}
	lastInsertId, err := res.LastInsertId()
	if err != nil {
		var empty Message
		return empty, err
	}
	// Instead of manually constructing a Message, fetch the complete record (which includes senderPicture)
	return db.GetMessageByID(uint64(lastInsertId))
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
func (db *appdbimpl) ForwardMessage(user User, convId, msgId, targetConvId uint64) (Message, error) {
	// Retrieve the original message using GetMessageByID.
	m, err := db.GetMessageByID(msgId)
	if err != nil {
		return Message{}, err
	}
	// Ensure the message belongs to the conversation specified by convId.
	if m.ConversationId != convId {
		return Message{}, errors.New("message does not belong to the specified conversation")
	}

	// Insert a new message in the target conversation with is_forwarded flagged.
	res, err := db.c.Exec("INSERT INTO messages(conversation_id, sender_id, content, format, state, is_forwarded) VALUES (?, ?, ?, ?, ?, ?)",
		targetConvId, user.Id, m.Content, m.Format, "Sent", 1)
	if err != nil {
		return Message{}, err
	}
	lastInsertId, err := res.LastInsertId()
	if err != nil {
		return Message{}, err
	}
	// Retrieve the new forwarded message.
	newMsg, err := db.GetMessageByID(uint64(lastInsertId))
	if err != nil {
		return Message{}, err
	}
	// Force the forwarded flag to be true if not set properly
	newMsg.IsForwarded = true
	return newMsg, nil
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
	SELECT 
	  m.id as id, 
	  m.conversation_id as conversationId, 
	  m.sender_id as senderId, 
	  u.username as senderName, 
	  u.profilePicture as senderPicture, 
	  m.content as content, 
	  m.format as format, 
	  m.state as state, 
	  m.timestamp as timestamp, 
	  m.reply_to as replyTo, 
	  m.is_forwarded as isForwarded
	FROM messages m
	JOIN users u ON m.sender_id = u.id
	WHERE m.id = ?
	`
	var isForwarded int
	err := db.c.QueryRow(query, msgId).Scan(
		&m.Id, &m.ConversationId, &m.SenderId, &m.SenderName, &m.SenderPicture,
		&m.Content, &m.Format, &m.State, &m.Timestamp, &replyTo, &isForwarded)
	if err != nil {
		return m, err
	}
	if replyTo.Valid {
		id := uint64(replyTo.Int64)
		m.ReplyToID = &id
	}
	m.IsForwarded = isForwarded != 0
	// Load reactions for the message.
	reactions, err := db.getMessageReactions(m.Id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return m, err
	}
	m.Reactions = reactions
	return m, nil
}

// MarkMessagesAsRead sets the state of all messages in a conversation that are not sent by the user to "Read".
func (db *appdbimpl) MarkMessagesAsRead(user User, convId uint64) error {
	_, err := db.c.Exec("UPDATE messages SET state = 'Read' WHERE conversation_id = ? AND sender_id != ? AND state != 'Read'", convId, user.Id)
	return err
}

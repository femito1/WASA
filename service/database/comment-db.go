package database

import (
	"errors"
)

func (db *appdbimpl) CommentMessage(user User, convId, msgId uint64, commentText string) (uint64, error) {
	// Optionally, verify that the message belongs to the conversation.
	var dummy int
	err := db.c.QueryRow("SELECT 1 FROM messages WHERE id = ? AND conversation_id = ?", msgId, convId).Scan(&dummy)
	if err != nil {
		return 0, err
	}
	res, err := db.c.Exec("INSERT INTO comments(message_id, user_id, commentText, senderName) VALUES (?, ?, ?, ?)", msgId, user.Id, commentText, user.Username)
	if err != nil {
		return 0, err
	}
	commentId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(commentId), nil
}

// DeleteComment removes a comment from a message.
func (db *appdbimpl) DeleteComment(user User, convId, msgId, commentId uint64) error {
	var dummy int
	err := db.c.QueryRow("SELECT 1 FROM comments WHERE id = ? AND message_id = ?", commentId, msgId).Scan(&dummy)
	if err != nil {
		return err
	}
	res, err := db.c.Exec("DELETE FROM comments WHERE id = ? AND message_id = ?", commentId, msgId)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("comment not found")
	}
	return nil
}

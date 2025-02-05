// File: database/contacts-db.go
package database

import (
	"errors"
	"fmt"
)

// AddContact adds a contact for the given user.
// It prevents a user from adding themselves.
func (db *appdbimpl) AddContact(userID, contactID uint64) error {
	if userID == contactID {
		return errors.New("cannot add yourself as a contact")
	}
	// Verify the contact exists.
	var tmp uint64
	err := db.c.QueryRow("SELECT id FROM users WHERE id = ?", contactID).Scan(&tmp)
	if err != nil {
		return fmt.Errorf("contact user not found: %w", err)
	}
	_, err = db.c.Exec("INSERT INTO contacts (user_id, contact_id) VALUES (?, ?)", userID, contactID)
	return err
}

// ListContacts returns all contacts for the given user.
func (db *appdbimpl) ListContacts(userID uint64) ([]User, error) {
	query := `
      SELECT u.id, u.username, u.profilePicture
      FROM users u
      JOIN contacts c ON u.id = c.contact_id
      WHERE c.user_id = ?
    `
	rows, err := db.c.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var contacts []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Id, &u.Username, &u.ProfilePicture); err != nil {
			return nil, err
		}
		contacts = append(contacts, u)
	}
	return contacts, nil
}

// RemoveContact removes a contact for the given user.
func (db *appdbimpl) RemoveContact(userID, contactID uint64) error {
	res, err := db.c.Exec("DELETE FROM contacts WHERE user_id = ? AND contact_id = ?", userID, contactID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("contact not found")
	}
	return nil
}

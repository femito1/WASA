package database

import (
	"database/sql"
	"errors"
)

// CreateUser creates a new user. If the username already exists, returns the existing user.
func (db *appdbimpl) CreateUser(u User) (User, error) {
	res, err := db.c.Exec("INSERT INTO users(username, profilePicture) VALUES (?, ?)", u.Username, u.ProfilePicture)
	if err != nil {
		// Likely the user already exists; try to fetch it.
		var existing User
		err2 := db.c.QueryRow("SELECT id, username, profilePicture FROM users WHERE username = ?", u.Username).
			Scan(&existing.Id, &existing.Username, &existing.ProfilePicture)
		if err2 != nil {
			if errors.Is(err2, sql.ErrNoRows) {
				return existing, ErrUserDoesNotExist
			}
			return existing, err2
		}
		return existing, nil
	}
	lastInsertId, err := res.LastInsertId()
	if err != nil {
		return u, err
	}
	u.Id = uint64(lastInsertId)
	return u, nil
}

// SetUsername updates the username of a user.
func (db *appdbimpl) SetUsername(u User, newUsername string) (User, error) {
	res, err := db.c.Exec("UPDATE users SET username = ? WHERE id = ? AND username = ?", newUsername, u.Id, u.Username)
	if err != nil {
		return u, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return u, err
	}
	if affected == 0 {
		return u, errors.New("no rows updated")
	}
	u.Username = newUsername
	return u, nil
}

// SetPhoto updates the profile picture of a user.
func (db *appdbimpl) SetPhoto(u User, newPic string) (User, error) {
	res, err := db.c.Exec("UPDATE users SET profilePicture = ? WHERE id = ?", newPic, u.Id)
	if err != nil {
		return u, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return u, err
	}
	if affected == 0 {
		return u, errors.New("no rows updated")
	}
	u.ProfilePicture = newPic
	return u, nil
}

// GetUserId fetches a user by username.
func (db *appdbimpl) GetUserId(username string) (User, error) {
	var user User
	err := db.c.QueryRow("SELECT id, username, profilePicture FROM users WHERE username = ?", username).
		Scan(&user.Id, &user.Username, &user.ProfilePicture)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, ErrUserDoesNotExist
		}
		return user, err
	}
	return user, nil
}

// CheckUserByUsername checks if a user exists by username.
func (db *appdbimpl) CheckUserByUsername(u User) (User, error) {
	var user User
	err := db.c.QueryRow("SELECT id, username, profilePicture FROM users WHERE username = ?", u.Username).
		Scan(&user.Id, &user.Username, &user.ProfilePicture)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, ErrUserDoesNotExist
		}
		return user, err
	}
	return user, nil
}

// CheckUserById checks if a user exists by id.
func (db *appdbimpl) CheckUserById(u User) (User, error) {
	var user User
	err := db.c.QueryRow("SELECT id, username, profilePicture FROM users WHERE id = ?", u.Id).
		Scan(&user.Id, &user.Username, &user.ProfilePicture)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, ErrUserDoesNotExist
		}
		return user, err
	}
	return user, nil
}

// CheckUser checks if a user exists by id and username.
func (db *appdbimpl) CheckUser(u User) (User, error) {
	var user User
	err := db.c.QueryRow("SELECT id, username, profilePicture FROM users WHERE id = ? AND username = ?", u.Id, u.Username).
		Scan(&user.Id, &user.Username, &user.ProfilePicture)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, ErrUserDoesNotExist
		}
		return user, err
	}
	return user, nil
}

// ListUsers lists all users or filters by name if provided.
func (db *appdbimpl) ListUsers(nameFilter string) ([]User, error) {
	var rows *sql.Rows
	var err error
	if nameFilter != "" {
		rows, err = db.c.Query("SELECT id, username, profilePicture FROM users WHERE username LIKE ?", "%"+nameFilter+"%")
	} else {
		rows, err = db.c.Query("SELECT id, username, profilePicture FROM users")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Id, &u.Username, &u.ProfilePicture); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// Example functions for get-name and set-name.
func (db *appdbimpl) GetName() (string, error) {
	var name string
	err := db.c.QueryRow("SELECT name FROM example_table WHERE id=1").Scan(&name)
	return name, err
}

func (db *appdbimpl) SetName(name string) error {
	_, err := db.c.Exec("INSERT OR REPLACE INTO example_table (id, name) VALUES (1, ?)", name)
	return err
}

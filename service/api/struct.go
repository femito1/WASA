package api

import (
	"github.com/femito1/WASA/service/database"
)

type User struct {
	Id       uint64 `json:"id"`
	Username string `json:"username"`
}

func (u *User) FromDatabase(user database.User) {
	u.Id = user.Id
	u.Username = user.Username
}

func (u *User) ToDatabase() database.User {
	return database.User{
		Id:       u.Id,
		Username: u.Username,
	}
}

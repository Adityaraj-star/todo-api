package store

import (
	"errors"

	"github.com/Adityaraj-star/todo-api/model"
)

var ErrUserExists = errors.New("username already taken")
var ErrUserNotFound = errors.New("user not found")

type UserStore interface {
	CreateUser(user model.User) (model.User, error)
	GetUserByUsername(username string) (model.User, error)
}
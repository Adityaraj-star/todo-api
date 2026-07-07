package store

import (
	"errors"

	"github.com/Adityaraj-star/todo-api/model"
)

var ErrNotFound = errors.New("todo not found")

type ListParams struct {
	Status string
	Limit  int
	Offset int
}

type TodoStore interface {
	GetAll(userID string, params ListParams) ([]model.Todo, error)
	GetByID(id, userID string) (model.Todo, error)
	Create(todo model.Todo) (model.Todo, error)
	Update(todo model.Todo) (model.Todo, error)
	Delete(id, userID string) error
}


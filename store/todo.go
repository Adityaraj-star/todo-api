package store

import (
	"sync"
	"errors"
	
	"github.com/Adityaraj-star/todo-api/model"
)

var ErrNotFound = errors.New("todo not found")

type TodoStore struct {
	todos 	map[string]model.Todo
	mu 		sync.Mutex
}

func NewTodoStore() *TodoStore {
	return &TodoStore{
		todos: make(map[string]model.Todo),
	}
}

func (r *TodoStore) GetAll() []model.Todo {
	r.mu.Lock()
	defer r.mu.Unlock()

	todos := []model.Todo{}

	for _, todo := range r.todos {
		todos = append(todos, todo)
	}

	return todos
}

func (r *TodoStore) GetByID(id string) (model.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	todo, ok := r.todos[id]

	if !ok {
		return model.Todo{}, ErrNotFound
	}
	return todo, nil
}

func (r *TodoStore) Create(todo model.Todo) model.Todo {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.todos[todo.ID] = todo

	return todo
}

func (r *TodoStore) Update(todo model.Todo) (model.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.todos[todo.ID]
	if !ok {
		return model.Todo{}, ErrNotFound
	}
	
	r.todos[todo.ID] = todo
	return todo, nil
}

func (r *TodoStore) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.todos[id]
	if !ok {
		return ErrNotFound
	}

	delete(r.todos, id)
	return nil
}
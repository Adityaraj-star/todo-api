package store

import (
	"sync"

	"github.com/Adityaraj-star/todo-api/model"
)

type MemoryTodoStore struct {
	todos map[string]model.Todo
	mu    sync.Mutex
}

func NewMemoryTodoStore() *MemoryTodoStore {
	return &MemoryTodoStore{
		todos: make(map[string]model.Todo),
	}
}

func (r *MemoryTodoStore) GetAll(userID string, params ListParams) ([]model.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	matched := []model.Todo{}
	for _, todo := range r.todos {
		if todo.UserID != userID {
			continue
		}
		if params.Status != "" && todo.Status != params.Status {
			continue
		}
		matched = append(matched, todo)
	}

	start := params.Offset
	if start > len(matched) {
		start = len(matched)
	}
	end := len(matched)
	if params.Limit > 0 && start+params.Limit < end {
		end = start + params.Limit
	}
	return matched[start:end], nil
}

func (r *MemoryTodoStore) GetByID(id, userID string) (model.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	todo, ok := r.todos[id]
	if !ok || todo.UserID != userID {
		return model.Todo{}, ErrNotFound
	}
	return todo, nil
}

func (r *MemoryTodoStore) Create(todo model.Todo) (model.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.todos[todo.ID] = todo
	return todo, nil
}

func (r *MemoryTodoStore) Update(todo model.Todo) (model.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.todos[todo.ID]
	if !ok || existing.UserID != todo.UserID {
		return model.Todo{}, ErrNotFound
	}
	r.todos[todo.ID] = todo
	return todo, nil
}

func (r *MemoryTodoStore) Delete(id, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.todos[id]
	if !ok || existing.UserID != userID {
		return ErrNotFound
	}
	delete(r.todos, id)
	return nil
}
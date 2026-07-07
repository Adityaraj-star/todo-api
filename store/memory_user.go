package store

import (
	"sync"

	"github.com/Adityaraj-star/todo-api/model"
)

type MemoryUserStore struct {
	users map[string]model.User
	mu    sync.Mutex
}

func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{
		users: make(map[string]model.User),
	}
}

func (s *MemoryUserStore) CreateUser(user model.User) (model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.Username]; exists {
		return model.User{}, ErrUserExists
	}
	s.users[user.Username] = user
	return user, nil
}

func (s *MemoryUserStore) GetUserByUsername(username string) (model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[username]
	if !ok {
		return model.User{}, ErrUserNotFound
	}
	return user, nil
}
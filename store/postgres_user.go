package store

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/Adityaraj-star/todo-api/model"
)

type PostgresUserStore struct {
	db *sql.DB
}

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{db: db}
}

func (s *PostgresUserStore) CreateUser(user model.User) (model.User, error) {
	_, err := s.db.Exec(
		`INSERT INTO users (id, username, password_hash, created_at) VALUES ($1, $2, $3, $4)`,
		user.ID, user.Username, user.PasswordHash, user.CreatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return model.User{}, ErrUserExists
		}
		return model.User{}, err
	}
	return user, nil
}

func (s *PostgresUserStore) GetUserByUsername(username string) (model.User, error) {
	var u model.User
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, created_at FROM users WHERE username = $1`,
		username,
	)
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	return u, nil
}
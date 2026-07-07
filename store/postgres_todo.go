package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Adityaraj-star/todo-api/model"
)

type PostgresTodoStore struct {
	db *sql.DB
}

func NewPostgresTodoStore(db *sql.DB) *PostgresTodoStore {
	return &PostgresTodoStore{db: db}
}

func (s *PostgresTodoStore) GetAll(userID string, params ListParams) ([]model.Todo, error) {
	query := `SELECT id, user_id, title, status, created_at FROM todos WHERE user_id = $1`
	args := []any{userID}

	if params.Status != "" {
		args = append(args, params.Status)
		query += fmt.Sprintf(" AND status = $%d", len(args))
	}

	query += " ORDER BY created_at DESC"

	if params.Limit > 0 {
		args = append(args, params.Limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	if params.Offset > 0 {
		args = append(args, params.Offset)
		query += fmt.Sprintf(" OFFSET $%d", len(args))
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []model.Todo{}
	for rows.Next() {
		var t model.Todo
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, rows.Err()
}

func (s *PostgresTodoStore) GetByID(id, userID string) (model.Todo, error) {
	var t model.Todo
	row := s.db.QueryRow(
		`SELECT id, user_id, title, status, created_at FROM todos WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.Status, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Todo{}, ErrNotFound
	}
	if err != nil {
		return model.Todo{}, err
	}
	return t, nil
}

func (s *PostgresTodoStore) Create(todo model.Todo) (model.Todo, error) {
	_, err := s.db.Exec(
		`INSERT INTO todos (id, user_id, title, status, created_at) VALUES ($1, $2, $3, $4, $5)`,
		todo.ID, todo.UserID, todo.Title, todo.Status, todo.CreatedAt,
	)
	if err != nil {
		return model.Todo{}, err
	}
	return todo, nil
}

func (s *PostgresTodoStore) Update(todo model.Todo) (model.Todo, error) {
	result, err := s.db.Exec(
		`UPDATE todos SET title = $1, status = $2 WHERE id = $3 AND user_id = $4`,
		todo.Title, todo.Status, todo.ID, todo.UserID,
	)
	if err != nil {
		return model.Todo{}, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return model.Todo{}, err
	}
	if rows == 0 {
		return model.Todo{}, ErrNotFound
	}
	return todo, nil
}

func (s *PostgresTodoStore) Delete(id, userID string) error {
	result, err := s.db.Exec(`DELETE FROM todos WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
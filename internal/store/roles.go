package store

import (
	"context"
	"database/sql"
	"errors"
)

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       int    `json:"level"`
	CreatedAt   string `json:"created_at"`
}

type RolesStore struct {
	db *sql.DB
}

func (s *RolesStore) GetByName(ctx context.Context, name string) (*Role, error) {

	query :=
		`
	SELECT id, name, description, level, created_at FROM roles WHERE name = $1;
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	role := &Role{}
	err := s.db.QueryRowContext(ctx, query, name).Scan(&role.ID, &role.Name, &role.Description, &role.Level, &role.CreatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return role, nil
}

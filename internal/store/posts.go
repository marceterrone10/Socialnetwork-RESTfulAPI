package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Comments  []Comment `json:"comments"`
}

type PostsStore struct {
	db *sql.DB
}

func (s *PostsStore) Create(ctx context.Context, post *Post) error { // se pasa contexto para que se pueda cancelar la operación si el contexto es cancelado
	query := `INSERT INTO posts (title, content, user_id, tags) 
	VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at;
	`

	err := s.db.QueryRowContext( // si pasamos contexto al la funcion, tenemos que usar QueryRowContext en lugar de QueryRow
		ctx,
		query,
		post.Title,
		post.Content,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostsStore) GetById(ctx context.Context, id int64) (*Post, error) {
	var post Post
	query := `SELECT id, title, content, user_id, tags, created_at, updated_at FROM posts WHERE id = $1;`

	err := s.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(&post.ID, &post.Title, &post.Content, &post.UserID, pq.Array(&post.Tags), &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (s *PostsStore) Delete(ctx context.Context, id int64) (*Post, error) {
	var post Post
	query := `DELETE FROM posts WHERE id = $1 RETURNING id, title, content, user_id, tags, created_at, updated_at;`

	err := s.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.UserID,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &post, nil
}

func (s *PostsStore) Update(ctx context.Context, post *Post) (*Post, error) {
	query := `
	UPDATE posts 
	SET title = $1, content = $2
	WHERE id = $3
	`
	_, err := s.db.ExecContext(ctx, query, post.Title, post.Content, post.ID)
	if err != nil {
		return nil, err
	}
	return post, nil

}

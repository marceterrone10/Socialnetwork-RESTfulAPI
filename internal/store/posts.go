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
	User      User      `json:"user"`
}

type PostWithMetadata struct {
	Post
	CommentCount int `json:"comment_count"`
}

type PostsStore struct {
	db *sql.DB
}

func (s *PostsStore) GetFeed(ctx context.Context, userId int64, fq PaginatedQuery) ([]*PostWithMetadata, error) {
	query := `
	SELECT p.id, p.title, p.content, p.user_id, p.tags, p.created_at, p.updated_at, COUNT(c.id) as comment_count, u.id as user_id, u.username, u.email
	FROM posts p
	LEFT JOIN comments c ON c.post_id = p.id
	JOIN users u ON u.id = p.user_id
	WHERE p.user_id = $1 OR p.user_id IN (SELECT follower_id FROM followers WHERE user_id = $1)
	GROUP BY p.id, u.id
	ORDER BY p.created_at ` + fq.Sort + `
	LIMIT $2 OFFSET $3;
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userId, fq.Limit, fq.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []*PostWithMetadata{}
	for rows.Next() {
		var post PostWithMetadata
		err :=
			rows.Scan(
				&post.Post.ID,
				&post.Post.Title,
				&post.Post.Content,
				&post.Post.UserID,
				pq.Array(&post.Post.Tags),
				&post.Post.CreatedAt,
				&post.Post.UpdatedAt,
				&post.CommentCount,
				&post.Post.User.ID,
				&post.Post.User.Username,
				&post.Post.User.Email,
			)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil

}

func (s *PostsStore) Create(ctx context.Context, post *Post) error { // se pasa contexto para que se pueda cancelar la operación si el contexto es cancelado
	query := `INSERT INTO posts (title, content, user_id, tags) 
	VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
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

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
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

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
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
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, post.Title, post.Content, post.ID)
	if err != nil {
		return nil, err
	}
	return post, nil

}

package store

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNotFound = errors.New("record not found")
)

type PostRepository interface { // aca vamos a tener las operaciones que vamos a hacer sobre los posts
	Create(context.Context, *Post) error
	GetById(context.Context, int64) (*Post, error)
	Delete(context.Context, int64) (*Post, error)
	Update(context.Context, *Post) (*Post, error)
}

type UserRepository interface { // aca vamos a tener las operaciones que vamos a hacer sobre los usuarios
	Create(context.Context, *User) error
}

type CommentRepository interface {
	GetByPostId(context.Context, int64) (*[]Comment, error)
}

type Storage struct { // inyección de dependencias de los repos
	Posts    PostRepository
	Users    UserRepository
	Comments CommentRepository
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:    &PostsStore{db},
		Users:    &UsersStore{db},
		Comments: &CommentsStore{db},
	}
}

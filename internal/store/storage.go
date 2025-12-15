package store

import (
	"context"
	"database/sql"
)

type PostRepository interface { // aca vamos a tener las operaciones que vamos a hacer sobre los posts
	Create(context.Context, *Post) error
}

type UserRepository interface { // aca vamos a tener las operaciones que vamos a hacer sobre los usuarios
	Create(context.Context, *User) error
}

type Storage struct { // inyección de dependencias de los repos
	Posts PostRepository
	Users UserRepository
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts: &PostsStore{db},
		Users: &UsersStore{db},
	}
}

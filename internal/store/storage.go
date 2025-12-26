package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound                        = errors.New("record not found")
	QueryTimeoutDuration time.Duration = 5 * time.Second
)

type PostRepository interface { // aca vamos a tener las operaciones que vamos a hacer sobre los posts
	Create(context.Context, *Post) error
	GetById(context.Context, int64) (*Post, error)
	Delete(context.Context, int64) (*Post, error)
	Update(context.Context, *Post) (*Post, error)
	GetFeed(context.Context, int64, PaginatedQuery) ([]*PostWithMetadata, error)
}

type UserRepository interface { // aca vamos a tener las operaciones que vamos a hacer sobre los usuarios
	Create(context.Context, *sql.Tx, *User) error
	GetById(context.Context, int64) (*User, error)
	GetByEmail(context.Context, string) (*User, error)
	CreateInvitation(ctx context.Context, user *User, token string, invitationExp time.Duration) error
	ActivateUser(ctx context.Context, token string) error
	Delete(ctx context.Context, userID int64) error
}

type CommentRepository interface {
	GetByPostId(context.Context, int64) (*[]Comment, error)
	Create(context.Context, *Comment) error
}

type FollowRepository interface {
	Follow(context.Context, int64, int64) error
	Unfollow(context.Context, int64, int64) error
}

type RoleRepository interface {
	GetByName(context.Context, string) (*Role, error)
}

type Storage struct { // inyección de dependencias de los repos
	Posts    PostRepository
	Users    UserRepository
	Comments CommentRepository
	Follows  FollowRepository
	Roles    RoleRepository
}

func NewStorage(db *sql.DB) Storage { // constructor del storage
	return Storage{
		Posts:    &PostsStore{db},
		Users:    &UsersStore{db},
		Comments: &CommentsStore{db},
		Follows:  &FollowsStore{db},
		Roles:    &RolesStore{db},
	}
}

// Funcion reutilizable para ejecutar transacciones en la base de datos
func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil) // comienza la transacción
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil { // ejecuta la función pasada como argumento
		_ = tx.Rollback() // si hay un error, cancela la transacción, devuelve el error
		return err
	}

	return tx.Commit() // commitea la transacción
}

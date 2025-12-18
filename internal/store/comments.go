package store

import (
	"context"
	"database/sql"
)

type Comment struct {
	ID        int64  `json:"id"`
	PostID    int64  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	User      User   `json:"user"`
}

type CommentsStore struct {
	db *sql.DB
}

func (s *CommentsStore) GetByPostId(ctx context.Context, postId int64) (*[]Comment, error) {
	query := `
	SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, users.id, users.username, users.email 
	FROM comments c
	JOIN users ON users.id = c.user_id
	WHERE c.post_id = $1
	ORDER BY c.created_at DESC;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	rows, err := s.db.QueryContext(ctx, query, postId) // variable rows para obtener las filas de la query que se ejecuta, se ejecuta conectandola a la DB y pasandole el contexto.
	if err != nil {
		return nil, err
	}
	defer rows.Close() // siempre se tienen que cerrar las filas para liberar recursos y no fugar la memoria

	comments := []Comment{} // comments es un slice de los comentarios que se vamos a tener de la query
	for rows.Next() {       // se va a ejecutar hasta que no haya más filas
		var c Comment                                                                                                        // c de tipo Comment para almacenar los datos de los comentarios de la fila
		c.User = User{}                                                                                                      // asignamos el usuario de tipo User para almacenar los datos del usuario de la fila
		err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.CreatedAt, &c.User.ID, &c.User.Username, &c.User.Email) // se scanean los datos de la fila y se asignan a las variables de c
		if err != nil {
			return nil, err
		}
		comments = append(comments, c) // Por ultimo se agrega el comentario a la slice de comentarios
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &comments, nil
}

// Los slices son estructuras de datos que se utilizan para almacenar una colección de elementos del mismo tipo.

func (s *CommentsStore) Create(ctx context.Context, comment *Comment) error {
	query :=
		`
	INSERT INTO comments (post_id, user_id, content) VALUES ($1, $2, $3) RETURNING id, created_at;
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := s.db.QueryRowContext(
		ctx,
		query,
		comment.PostID,
		comment.UserID,
		comment.Content,
	).Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		return err
	}

	userQuery :=
		`
	SELECT id, username, email FROM users WHERE id = $1;
	`
	comment.User = User{}
	err = s.db.QueryRowContext(
		ctx,
		userQuery,
		comment.UserID,
	).Scan(&comment.User.ID, &comment.User.Username, &comment.User.Email)
	if err != nil {
		return err
	}
	return nil

}

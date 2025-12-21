package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/marceterrone10/social/internal/store"
)

type postKey string

const postCtx postKey = "post" // clave para el contexto del post

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

// CreatePost godoc
//
//	@Summary		Create a new post
//	@Description	Create a new post with a title, content, and tags
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreatePostPayload	true	"Post payload"
//	@Success		201		{object}	store.Post			"Post created successfully"
//	@Failure		400		{object}	error				"Bad request"
//	@Failure		500		{object}	error				"Internal server error"
//	@Router			/posts [post]
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {

	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	} // leemos el payload del request y se parsea

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	} // validamos el payload

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  1,
	}

	ctx := r.Context()

	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	} // creamos el post en la base de datos

	if err := app.writeResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	} // escribimos el post creado en el response
}

// GetPost godoc
//
//	@Summary		Get a post by ID
//	@Description	Get a post by ID
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int			true	"Post ID"
//	@Success		200	{object}	store.Post	"Post found"
//	@Failure		400	{object}	error		"Bad request"
//	@Failure		404	{object}	error		"Post not found"
//	@Failure		500	{object}	error		"Internal server error"
//	@Router			/posts/{id} [get]
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r.Context())

	comments, err := app.store.Comments.GetByPostId(r.Context(), post.ID) // obtenemos los comentarios del post
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = *comments // asignamos los comentarios al post

	if err := app.writeResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	} // escribimos el post en el response

}

// DeletePost godoc
//
//	@Summary		Delete a post by ID
//	@Description	Delete a post by ID
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Post ID"
//	@Success		204	"Post deleted successfully"
//	@Failure		400	{object}	error	"Bad request"
//	@Failure		404	{object}	error	"Post not found"
//	@Failure		500	{object}	error	"Internal server error"
//	@Router			/posts/{id} [delete]
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	} // parseamos el id del post
	ctx := r.Context()

	_, err = app.store.Posts.Delete(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	} // eliminamos el post

	w.WriteHeader(http.StatusNoContent)
}

type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

// UpdatePost godoc
//
//	@Summary		Update a post by ID
//	@Description	Update a post by ID
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Post ID"
//	@Param			payload	body		UpdatePostPayload	true	"Post payload"
//	@Success		200		{object}	store.Post			"Post updated successfully"
//	@Failure		400		{object}	error				"Bad request"
//	@Failure		500		{object}	error				"Internal server error"
//	@Router			/posts/{id} [patch]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r.Context())

	var payload UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Title != nil {
		post.Title = *payload.Title
	}

	post, err := app.store.Posts.Update(r.Context(), post)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.badRequestError(w, r, err)
			return
		} // parseamos el id del post

		ctx := r.Context()

		post, err := app.store.Posts.GetById(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		} // obtenemos el post por id

		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(ctx context.Context) *store.Post {
	post, _ := ctx.Value(postCtx).(*store.Post)

	return post
} // obtenemos el post del contexto

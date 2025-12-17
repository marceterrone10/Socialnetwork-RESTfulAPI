package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/marceterrone10/social/internal/store"
)

type CreatePostPayload struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {

	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	} // leemos el payload del request y se parsea

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  1,
	}

	ctx := r.Context()

	if err := app.store.Posts.Create(ctx, post); err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	} // creamos el post en la base de datos

	if err := writeJSON(w, http.StatusCreated, post); err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	} // escribimos el post creado en el response
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	} // parseamos el id del post

	ctx := r.Context()

	post, err := app.store.Posts.GetById(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			errorJSON(w, http.StatusNotFound, err.Error())
		default:
			errorJSON(w, http.StatusInternalServerError, err.Error())
		}
		return
	} // obtenemos el post por id

	if err := writeJSON(w, http.StatusOK, post); err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	} // escribimos el post en el response

}
